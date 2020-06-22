## 大文件上传

### 0、项目源码地址

前端基于 vue-simple-uploader （感谢这个大佬）实现： https://github.com/simple-uploader/vue-uploader/blob/master/README_zh-CN.md

vue-simple-uploader底层封装了uploader.js :    https://github.com/simple-uploader/Uploader/blob/develop/README_zh-CN.md

上面两个项目是在探究大文件上传过程中编写的Demo

>  他只是个demo，所以不再去死抠里面的细节和边界值了。不过起码的参考意义还是有的。

### 1、如何唯一标示一个文件？

文件的信息后端会存储在mysql数据库表中。

在上传之前，前端通过  spark-md.js 计算文件的md5值以此去唯一的标示一个文件。

sprk-md.js 地址：

README.md中有sprk-md.js的使用demo，可以去看看。



### 2、断点续传是如何实现的？

断点续传可以实现这样的功能，比如RD上传200M的文件，当RD上传完199M时，断网了，有了断点续传的功能，我们允许RD再次上传时，能从第199M的位置重新上传。

实现原理：

实现断点续传的前提是，大文件切片上传。然后前端得问后端哪些chunk曾经上传过，让前端跳过这些上传过的chunk就好了。

前端的上传器（uploader.js）在上传时会先发送一个GET请求，这个请求不会携带任何chunk数据，作用就是向后端询问哪些chunk曾经上传过。 后端会将这些数据保存在mysql数据库表中。比如按这种格式：`1:2:3:5`表示，曾经上传过的分片有1，2，3，5。第四片没有被上传，前端会跳过1，2，3，5。 仅仅会将第四个chunk发送给后端。


### 3、秒传是如何实现的？

秒传实现的功能是：当RD重复上传一份相同的文件时，除了第一次上传会正常发送上传请求后，其他的上传都会跳过真正的上传，直接显示秒成功。

实现方式：

后端存储着当前文件的相关信息。为了实现秒传，我们需要搞一个字段（isUploaded）表示当前md5对应的文件是否曾经上传过。 后端在处理 前端的上传器（uploader.js）发送的第一个GET请求时，会将这个字段发送给前端，比如 isUploaded = true。前端看到这个信息后，直接跳过上传，显示上传成功。



### 4、上传暂停是如何实现的？

上传的暂停：并不是去暂停一个已经发送出去的正在进行数据传输的http请求～  

而是暂停发送起发送下一个http请求。

就我们的项目而言，因为我们的文件本来就是先切片，对于我们来说，暂停文件的上传，本质上就是暂停发送下一个chunk。



### 5、前端上传并发数是多少？

前端的uploader.js中默认会三条线程启动并发上传，前端会在同一时刻并发 发送3个chunk，后端就会相应的为每个请求开启三个协程处理上传的过来的chunk。

在我们的项目中，会将前端并发数调整成了1。原因如下：

我们的项目中考虑到了断点续传的实现，后端需要记录下曾经上传过哪些切片。（这个记录在mysql的数据库表中，以 ”1:2:3:4:5“ ）这种格式记录。

Mysql5.7默认的存储引擎是innoDB，默认的隔离级别是RR。如果我们将前端的并发数调大，就会出现下面的异常情况：



```go
1. goroutine1 获取开启事物，读取当前上传到记录是 1:2 （未提交事物）
2. goroutine1 在现有的记录上加上自己处理的分片3，并和现有的1:2拼接在一起成1:2:3 （未提交事物）
3. goroutine2 获取开启事物，（因为RR，所以它读不到1:2:3）读取当前上传到记录是 1:2 （未提交事物）
4. goroutine1 提交事物，将1:2:3写回到mysql
5. goroutine2 在现有的记录上加上自己处理的分片4，并和现有的1:2拼接在一起成1:2:4 （提交事物）
```



可以看到，如果前端并发上传，后端就会出现分片丢失的问题。 故前端将并发数置为1。

### 6、单个chunk上传失败怎么办？

**前端会重传chunk？**

由于网络问题，或者时后端处理chunk时出现的其他未知的错误，会导致chunk上传失败。

uploaded.js 中有如下的配置项,  每次uploader.js 在上传每一个切片实际上都是在发送一次post请求，后端根据这个post请求是会给前端一个状态吗。 uploader.js 就是根据这个状态码去判断是失败了还是成功了，如果失败了就会重新发送这个上传的请求。

那uploader.js是如何知道有哪些状态吗是它应该重传chunk的标记呢？ 看看下面uploader.js需要的options 就明白了，其中的`permantErrors`中配置的状态码标示：当遇到这个状态码时整个上传直接失败～

`successStatuses`中配置的状态码表示chunk是上传成功的～。 其他的状态吗uploader.js 就会任务chunk上传的有问题，于是重新上传～

```js
        options: {
          target: 'http://localhost:8081/file/upload',
          maxChunkRetries: 3,
          permanentErrors:[502], // 永久性的上传失败～，会认为整个文件都上传失败了
          successStatuses:[200], // 当前chunk上传成功后的状态吗
          ...
        }
```



### 7、超过重传次数后，怎么办？

比如我们设置出错后重传的次数为3，那么无论当前分片是第几片，整个文件的上传状态被标记为false，这就意味着会终止所有的上传。

肯定不会出现这种情况：chunk1重传3次后失败了，chunk2还能再去上传，这样的话数据肯定不一致了。



### 8、如何控制上传多大的文件？

目前了解到nginx端的限制上单次上传不能超过1M。

前端会对大文件进行切片突破nginx的限制。

```js
        options: {
          target: 'http://localhost:8081/file/upload',
          chunkSize: 512000, // 单次上传 512KB 
        }     
```

如果后续和nginx负责的同学达成一致，可以把这个值进行调整。前端可以后续将这个chunk的阈值加大。



### 9、如何保证上传文件的百分百正确？

在上传文件前，前端会计算出当前RD选择的这个文件的 md5 值。

当后端检测到所有的分片全部上传完毕，这时会merge所有分片汇聚成单个文件。计算这个文件的md5 同 RD在前端提供的文件的md5值比对。 比对结果一致说明RD正确的完成了上传。结果不一致，说明文件上传失败了～返回给前端任务失败，提示RD重新上传。





### 10、其他细节问题：

#### 如何判断文件上传失败了，给RD展示红色？

#### 如何控制上传什么类型的文件？

#### 如何控制不能上传空文件？



上面说过了，当 uploader.js 遇到了`permanentErrors`这种状态码时会认为文件上传失败了。

前端想在上传失败后，将进度条转换成红色，其实改一下CSS样式就好了，问题就在于，根据什么去修改？在哪里去修改？

前端会将每一个file封装成一个组件：如下图中的files就是file的集合

![image-20200621214536333](https://img2020.cnblogs.com/blog/1496926/202006/1496926-20200622231557172-2130180178.png)

整个的fileList会将会被渲染成下面这样。

![image-20200621214718940](https://img2020.cnblogs.com/blog/1496926/202006/1496926-20200622231557931-318079909.png)

---

我们上传的文件会被封装成file对象，这个对象中会有个配置参数, 比如它会长下面这样。

```js
     options: {
        target: 'http://localhost:8081/file/upload',
        statusText: {
          success: '上传成功',
          error: '上传出错，请重试',
          typeError: '暂不支持上传您添加的文件格式',
          uploading: '上传中',
          emptyError:'不能上传空文件',
          paused: '请确认文件后点击上传',
          waiting: '等待中'
        }
      }
    },
```

我们将上面的配置添加给Uploader.js

```go
      const uploader = new Uploader(this.options)
```

在file组件中有如下计算属性的，分别是status和statusText

```js
    computed: {
      // 计算出一个状态信息
      status () {
        const isUploading = this.isUploading // 是否正在上传
        const isComplete = this.isComplete // 是否已经上传完成
        const isError = this.error // 是否出错了
        const isTypeError = this.typeError // 是否出错了
        const paused = this.paused // 是否暂停了
        const isEmpty = this.emptyError // 是否暂停了
        // 哪个属性先不为空，就返回哪个属性
        if (isComplete) {
          return 'success'
        } else if (isError) {
          return 'error'
        } else if (isUploading) {
          return 'uploading'
        } else if (isTypeError) {
          return 'typeError'
        } else if (isEmpty) {
          return 'emptyError'
        } else if (paused) {
          return 'paused'
        } else {
          return 'waiting'
        }
      },
      // 状态文本提示信息
      statusText () {
        // 获取到计算出的status属性（相当于是个key，具体的值在下面的fileStatusText中获取到）
        const status = this.status
        // 从file的uploader对象中获取到 fileStatusText，也就是用自己定义的名字
        const fileStatusText = this.file.uploader.fileStatusText
        let txt = status
        if (typeof fileStatusText === 'function') {
          txt = fileStatusText(status, this.response)
        } else {
          txt = fileStatusText[status]
        }
        return txt || status
      },
    },
```

status绑定在html上

```html
	<div class="uploader-file" :status="status">
```

对应的CSS样式入下：

```css
  .uploader-file[status="error"] .uploader-file-progress {
    background: #ffe0e0;
  }
```

综上：有了上面代码的编写，我们可以直接像下面这样控制就好了

```js
  file.typeError = true // 表示文件的类型不符合我们的预期，不允许RD上传
  file.error = true // 表示文件上传失败了
  file.emptyError = true // 表示文件为空，不允许上传
```



### 11、后端数据库表设计

```bash
CREATE TABLE `file_upload_detail` (                                                                               
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',                                                           
  `username` varchar(64) NOT NULL COMMENT '上传文件的用户账号',                                                            
  `file_name` varchar(64) NOT NULL COMMENT '上传文件名',                                                               
  `md5` varchar(255) NOT NULL COMMENT '上传文件的MD5值',                                                                
  `is_uploaded` int(11) DEFAULT '0' COMMENT '是否完整上传过 \n0：否\n1：是',                                                 
  `has_been_uploaded` varchar(1024) DEFAULT NULL COMMENT '曾经上传过的分片号',                                             
  `url` varchar(255) DEFAULT NULL COMMENT 'bos中的url，或者是本机的url地址',                                                 
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP  COMMENT '本条记录创建时间',     
  `update_time` timestamp NULL DEFAULT NULL  COMMENT '本条记录更新时间',                                                  
  `total_chunks` int(11) DEFAULT NULL COMMENT '文件的总分片数',                                                          
  PRIMARY KEY (`id`)                                                                                              
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8                                                             
```



### 12、关于什么时候mergechunk

在我本文中给出的demo中，merge是后端处理完成所有的chunk后，像前端返回 merge=1，这个表示来实现的。

前端拿着这个字段去发送/merge请求去合并所有的chunk。 

**值得注意的地方是：这个请求是在uploader.js认为所有的分片全部成功上传后，在单个文件成功上传的回调中执行的**。我想了一下，感觉这么搞其实不太友好，万一merge的过程中失败了，或者是某个chunk丢失了，chunk中的数据缺失，最终merge的产物的md5值其实并不等于原文件。当这种情况发生的时候，其实上传是失败的。但是后端既然告诉uploader.js 可以合并了，说明后端认为任务是成功的。vue-simple-uploader也会觉得任务是成功的，前端段展示绿色的上传成功给用户看～， 这么看来，整个过程其实控制的不太好～

我现在的实现：直接去掉merge请求，前端1条线程发送请求，将chunk依次发送到后端。后端检测到所有的chunk都上传过来后主动merge，merge完成后马上校验文件的md5值是否符合预期。这个处理过程在上传最后一个chunk的请求中进行，因此可以实现的控制前端上传成功还是失败的样式～
