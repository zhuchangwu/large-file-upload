<template>
  <div>
    <uploader :options="options"
              :file-status-text="statusText"
              class="uploader-example"
              ref="uploader"
              @file-complete="fileComplete"
              @complete="complete"
    >
    </uploader>
  </div>

</template>

<script>
  export default {
    data () {
      return {
        options: {
          target: 'http://localhost:8081/file/upload',
          simultaneousUploads:1,
          chunkSize: 1024000,
          maxChunkRetries: 3,
          permanentErrors:[502], // 永久性的上传失败～，会认为整个文件都上传失败了
          successStatuses:[200], // 当前chunk上传成功后的状态吗
          // 会发送与服务器进行分片校验的请求，第一个get请求就是该请求；后面的每一个post请求都是上传分片的请求
          checkChunkUploadedByResponse: function (chunk, response) {
            console.log("123")
            response = JSON.parse(response)
            console.log('checkChunkUploadedByResponse() chunk ', chunk)
            console.log('checkChunkUploadedByResponse() response ', response)
            // 秒传～
            if (response.isUploaded === 1) {
              console.log("秒传～")
              return true
            }
            // 断点续传~
            if (response.hasBeenUploaded) {
              var hasBeenUploaded = response.hasBeenUploaded.toString().split(':')
              console.log('hasBeenUploaded: ', hasBeenUploaded)
              for (var i = 0; i < hasBeenUploaded.length; i++) {
                if ((hasBeenUploaded[i] - 1) === chunk.offset) {
                  console.log('断点续传～')
                  return true
                }
              }
            }
            return false
          },
          // 开启服务器的分片校验
          // 在这次校验中我们向后端获取到曾经上传过到分片信息，进而实现断点续传到功能
          testChunks: true
        },
        /* 添加的额外属性 */
        attrs: {
          accept: 'image/*'
        },
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
    methods: {
      // fileList上传完毕后回调
      complete () {
        console.log('complete', arguments)
      },
      // 单个文件（文件夹）上传后都会回调
      fileComplete () {
        console.log('file complete', arguments)
      }
    },
    mounted () {
      this.$nextTick(() => {
        window.uploader = this.$refs.uploader.uploader
      })
    }
  }
</script>

<style>
  .uploader-example {
    width: 880px;
    padding: 15px;
    margin: 40px auto 0;
    font-size: 12px;
    box-shadow: 0 0 10px rgba(0, 0, 0, .4);
  }

  .uploader-example .uploader-btn {
    margin-right: 4px;
  }

  .uploader-example .uploader-list {
    max-height: 440px;
    overflow: auto;
    overflow-x: hidden;
    overflow-y: auto;
  }
</style>
