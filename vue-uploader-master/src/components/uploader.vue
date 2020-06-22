<template>
  <!--uploader组件-->
  <div class="uploader">
    <!--它里面只有一个插槽-->
    <slot :files="files" :file-list="fileList" :started="started">
      <!--如果浏览器不支持html5api，显示这个表前-->
      <uploader-unsupport></uploader-unsupport>
      <!-- 上传文件 -->
      <uploader-btn>选择文件</uploader-btn>
      <!-- 选择的文件列表 -->
      <uploader-list></uploader-list>
    </slot>
  </div>
</template>

<script>
  //
  // 当前组件Name：uploader （就是这个这个大佬在github中推出的 vue-simple-uploader）
  // 如下，vue-simple-uploader在项目中的名字就叫做 uploader
  // uploader中有6个子组件
  //   1：上传的按钮： UploaderBtn
  //   2：上传的拖拽框： UploaderDrop
  //   3：浏览器不支持html5API时的提示框： UploaderUnsupport
  //   4：用来展示将要上传的所有的文件列表的组件： UploaderList
  //   5：上传单个文件的组件： UploaderFile
  //   6：批量上传多个文件的组件： UploaderFiles， 它实际上是多个UploaderFile的集合
  //   -
  //   根据业务：我要做单个文件的上传，只需要上面的1、3、5就够了～
  //
  import Uploader from 'simple-uploader.js'
  import {kebabCase} from '../common/utils'
  import SparkMD5 from '../../example/spark-md5'
  import UploaderBtn from './btn.vue'
  import UploaderDrop from './drop.vue'
  import UploaderUnsupport from './unsupport.vue'
  import UploaderList from './list.vue'
  import UploaderFiles from './files.vue'
  import UploaderFile from './file.vue'

  // 发送merge请求
  import axios from 'axios'
  import qs from 'qs'

  /* 定义常量 */
  /* 组件名 */
  const COMPONENT_NAME = 'uploader'
  /* 三个事件名 */
  const FILE_ADDED_EVENT = 'fileAdded'
  const UPLOAD_START_EVENT = 'uploadStart'

  export default {
    // vue-simple-uploader的名字在项目中就叫做 uploader
    name: COMPONENT_NAME,
    /* 将组件映射成标签 */
    components: {
      UploaderBtn,
      UploaderDrop,
      UploaderUnsupport,
      UploaderList,
      UploaderFiles,
      UploaderFile
    },
    provide () {
      return {
        uploader: this
      }
    },
    /* 接收父组件给传递过来的属性 */
    props: {
      /* 属性 */
      options: {
        type: Object,
        default () {
          return {}
        }
      },
      /* 是否自动开始上传 */
      autoStart: {
        type: Boolean,
        default: false // false标示用户选好文件后，不会立即自动上传
      },
      /* 被上传的文件所处的状态 */
      fileStatusText: {
        type: [Object, Function],
        default () {
          return {
            success: 'success',
            error: 'error',
            uploading: 'uploading',
            paused: 'paused',
            waiting: 'waiting'
          }
        }
      }
    },
    data () {
      return {
        started: false,
        files: [],
        fileList: []
      }
    },
    methods: {
      /* 开始上传 */
      uploadStart () {
        // 开始上传前判断文件的类型
        this.started = true
      },
      /* 这个事件一般用作文件校验，如果说返回了 false，那么这个文件就会被忽略，不会添加到文件上传列表中。 */
      fileAdded (file) {
        console.log('file: ', file)
        // 限制文件的类型
        var name = file.name.toString()
        var surfix = name.substring(name.lastIndexOf('.') + 1, name.length)
        if (surfix !== 'dmg') {
          file.typeError = true
          return
        }
        // 校验文件得有内容，不能上传空文件
        if (file.chunks[0].endByte === 0) {
          file.emptyError = true
          return
        }

        // 暂停文件相关的操作
        file.pause()
        // 计算文件的md5
        this.computeMD5(file)
        this.$emit(kebabCase(FILE_ADDED_EVENT), file)
        if (file.ignored) {
          // is ignored, filter it
          return false
        }
        console.log('fileAdded')
      },
      computeMD5: function (file) {
        let time = new Date().getTime()
        let blobSlice = File.prototype.slice || File.prototype.mozSlice || File.prototype.webkitSlice
        let currentChunk = 0
        const chunkSize = 10 * 1024 * 1000
        let chunks = Math.ceil(file.size / chunkSize)
        let spark = new SparkMD5.ArrayBuffer()
        let fileReader = new FileReader()
        // 暂停对文件对上传
        file.pause()
        loadNext()
        fileReader.onload = (e => {
          // 分片读取文件追加进spark中
          spark.append(e.target.result)
          // 如果当前对分片数小于总分片数，我们就继续上传
          if (currentChunk < chunks) {
            currentChunk++
            loadNext()
            // 实时展示MD5的计算进度
            this.$nextTick(() => {
              // $(`.myStatus_${file.id}`).text('校验MD5 ' + ((currentChunk / chunks) * 100).toFixed(0) + '%')
            })
          } else {
            // 当前分片数已经不小于总分片数，说明上传完了～
            let md5 = spark.end()
            this.computeMD5Success(md5, file)
            console.log('MD5计算完毕：' + file.name + 'MD5： ' + md5 + '分片： ' + chunks + '大小:  ' + file.size + '用时：  ' + (new Date().getTime() - time))
          }
        })
        // FileReader读取不到文件时报错～
        fileReader.onerror = function () {
          alert('文件读取出错，请确认～')
          file.cancel()
        }

        // 计算出下一个用于截取文件的开始和结束的位置
        function loadNext () {
          let start = currentChunk * chunkSize
          let end = ((start + chunkSize) >= file.size) ? file.size : start + chunkSize
          fileReader.readAsArrayBuffer(blobSlice.call(file.file, start, end))
        }
      },
      // 计算md5成功后回调
      computeMD5Success (md5, file) {
        // 将自定义参数直接加载uploader实例的opts上
        Object.assign(this.uploader.opts, {
          query: {
            ...this.params
          }
        })
        file.uniqueIdentifier = md5
        // file.resume()
      },
      /* 取消上传的事件 */
      /* 一个文件（文件夹）被移除出待上传的文件列表中的事件 */
      fileRemoved (file) {
        this.files = this.uploader.files
        this.fileList = this.uploader.fileList
        console.log('点击了取消上传: ', file.name)
      },
      /* 和 filesAdded 类似，但是是文件已经加入到上传列表中，一般用来开始整个的上传。 */
      // 一般我们在这个函数中开始整体的上传操作
      filesSubmitted (files, fileList) {
        this.files = this.uploader.files
        this.fileList = this.uploader.fileList
        // 判断是否自动开始上传文件
        if (this.autoStart) {
          this.uploader.upload()
        }
        console.log('用户选好了文件～')
      },
      // 单个文件上传成功后回调～
       fileSuccess (rootFile, file, message, chunk) {
        let res = JSON.parse(message)
        // 服务器自定义的错误（即虽返回200，但是是错误的情况），这种错误是Uploader无法拦截的
        // 向后端发送请求merge文件
        /**
         *
         *  fileSuccess() message {
         *    "isUploaded": 0,
         *    "hasBeenUploaded": "",
         *    "merge": 1,
         *    "msg": "",
         *    "url": ""
         *    }
         */
        // 发送merge请求
        if (res.merge === 1) {
           this.merge(file)
        }
      },
       // 合并文件
       merge(file){
        axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded'
        axios.post(
          'http://localhost:8081/file/merge',
          qs.stringify({
            fileName: file.name,
            identifier: file.uniqueIdentifier
          })
        ).then((res) => {
          console.log('最后一步：后端给响应了～～～')
          console.log(res)
          //todo 上传成功～ 保存url
          //todo 上传失败～ 给前端提示， 最直观的就是修改样式为flase
          file.error = true
          console.log('最后一步：后端给响应了～～～')
        })
      },
      allEvent (...args) {
        const name = args[0]
        // 事件的标示
        const EVENTSMAP = {
          [FILE_ADDED_EVENT]: true, // 单个文件添加事件
          [UPLOAD_START_EVENT]: 'uploadStart' // 开始上传的事件
        }
        const handler = EVENTSMAP[name]
        if (handler) {
          if (handler === true) {
            return
          }
          this[handler].apply(this, args.slice(1))
        }
        args[0] = kebabCase(name)
        this.$emit.apply(this, args)
      }
    },
    //  vue-simple-uploader是对uploader.js的封装
    //  后续在实现大文件的上传，断点续传，妙传的功能时，需要注册的一些事件的回调 都在下面的created方法中注册～
    // 组件初始化时回调
    created () {
      this.options.initialPaused = !this.autoStart
      // 创建 uploader.js的实例
      // uploader的github地址：https://github.com/simple-uploader/Uploader/blob/develop/README_zh-CN.md#%E4%BA%8B%E4%BB%B6
      const uploader = new Uploader(this.options)
      // 将uploader.js的实例赋值给当前的vue-simple-uploader上传器
      this.uploader = uploader
      // 状态的赋值
      this.uploader.fileStatusText = this.fileStatusText
      // 注册事件，在底层使用的依然是upload.js, 需要我们将vue-simple-uploader中的事件都绑定在 uploader.js上
      // 下面就是为uploader.js 绑定上一些监听事件
      // 比如： 添加 单个文件的事件  fileAdded
      // 比如： 单个文件上传 成功 的事件  fileSuccess
      // 比如： 跟下的单个文件、文件夹上传 完成 的事件  fileComplete
      // 比如： 某个文件上传失败的事件 fileError
      // 告诉uploader.js我们有一共有哪些事件
      uploader.on('catchAll', this.allEvent)
      // 具体的为这些事件添加指定的回调函数
      uploader.on(FILE_ADDED_EVENT, this.fileAdded)
      // 一个文件（文件夹）被移除
      uploader.on('fileRemoved', this.fileRemoved)
      // 单个文件上传成功
      uploader.on('fileSuccess', this.fileSuccess)
      // 和 filesAdded 类似，但是是文件已经加入到上传列表中，一般用来开始整个的上传。
      uploader.on('filesSubmitted', this.filesSubmitted)
    },
    // 在 vue-simple-uploader组件销毁前调用，收尾～
    destroyed () {
      const uploader = this.uploader
      uploader.off('catchAll', this.allEvent)
      uploader.off(FILE_ADDED_EVENT, this.fileAdded)
      uploader.off('fileRemoved', this.fileRemoved)
      uploader.off('filesSubmitted', this.filesSubmitted)
      this.uploader = null
    }
  }
</script>

<style>
  .uploader {
    position: relative;
  }
</style>
