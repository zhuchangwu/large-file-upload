<template>
  <div class="uploader-file" :status="status">
    <slot
      :file="file"
      :list="list"
      :status="status"
      :emptyError="emptyError"
      :paused="paused"
      :error="error"
      :typeError="typeError"
      :response="response"
      :average-speed="averageSpeed"
      :formated-average-speed="formatedAverageSpeed"
      :current-speed="currentSpeed"
      :is-complete="isComplete"
      :is-uploading="isUploading"
      :size="size"
      :formated-size="formatedSize"
      :uploaded-size="uploadedSize"
      :progress="progress"
      :progress-style="progressStyle"
      :progressing-class="progressingClass"
      :time-remaining="timeRemaining"
      :formated-time-remaining="formatedTimeRemaining"
      :type="type"
      :extension="extension"
      :file-category="fileCategory"
      :md5Process="md5Process"
    >
      <!--上传的进度条-->
      <div class="uploader-file-progress" :class="progressingClass" :style="progressStyle"></div>
      <!--上传的文件信息-->
      <div class="uploader-file-info">
        <!--根据文件的类型，显示不同的icon、以及文件名-->
        <div class="uploader-file-name">
          <i class="uploader-file-icon" :icon="fileCategory"></i>
          {{file.name}}
        </div>
        <!-- 上传的文件大小 -->
        <div class="uploader-file-size">{{formatedSize}}</div>
        <!-- 计算md5的进度 -->
        <div class="uploader-file-meta">{{md5Process}}</div>
        <!-- 上传的状态 -->
        <div class="uploader-file-status">
          <!--这个status是个计算属性-->
          <span v-show="status !== 'uploading'">{{statusText}}</span>
          <span v-show="status === 'uploading'">
            <span>{{progressStyle.progress}}</span>
            <em>{{formatedAverageSpeed}}</em>
            <i>{{formatedTimeRemaining}}</i>
          </span>
        </div>
        <!-- 操作的按钮 -->
        <div class="uploader-file-actions">
          <span class="uploader-file-pause" @click="pause"></span>
          <span class="uploader-file-resume" @click="resume">️</span>
          <span class="uploader-file-retry" @click="retry"></span>
          <span class="uploader-file-remove" @click="remove"></span>
        </div>
      </div>
    </slot>
  </div>
</template>

<script>
  import Uploader from 'simple-uploader.js'
  import events from '../common/file-events'
  import {secondsToStr} from '../common/utils'

  const COMPONENT_NAME = 'uploader-file'

  export default {
    name: COMPONENT_NAME,
    props: {
      // 上传的文件实例
      file: {
        type: Object,
        default () {
          return {}
        }
      },
      // 是否在uploaderList组件中使用
      list: {
        type: Boolean,
        default: false
      }
    },
    data () {
      return {
        response: null,
        paused: false, // 是否暂停了
        error: false, // 是否出错了
        averageSpeed: 0, // 平均上传速度 3kb/s
        currentSpeed: 0, // 当前上传速度，单位字节每秒
        isComplete: false, // 是否完成上传了
        isUploading: false, // 是否正在上传
        size: 0, // 文件、文件夹的大小
        formatedSize: '', // 格式化后，文件或者文件夹的大小
        uploadedSize: 0, // 已上传的大小
        progress: 0, // 介于0～1的小数，上传进度
        timeRemaining: 0, // 预估的剩余的时间，单位秒
        type: '', // 文件的类型
        extension: '', // 文件的后缀名
        progressingClass: '', // 进度条的class属性
        md5Process: '', // 进度条的class属性
        typeError: false,//文件类型错误
        emptyError: false //文件为空
      }
    },
    computed: {
      // 通过计算属性，创建fileCategory属性
      // 文件的类型如： `folder`, `document`, `video`, `audio`, `image`, `unknown`
      fileCategory () {
        const extension = this.extension // 文件的后缀
        const isFolder = this.file.isFolder // 是否是文件夹
        let type = isFolder ? 'folder' : 'unknown'
        const categoryMap = this.file.uploader.opts.categoryMap
        // 所有的类型被放在一个map，
        const typeMap = categoryMap || {
          image: ['gif', 'jpg', 'jpeg', 'png', 'bmp', 'webp'],
          video: ['mp4', 'm3u8', 'rmvb', 'avi', 'swf', '3gp', 'mkv', 'flv'],
          audio: ['mp3', 'wav', 'wma', 'ogg', 'aac', 'flac'],
          document: ['doc', 'sql', 'txt', 'docx', 'pages', 'epub', 'pdf', 'numbers', 'csv', 'xls', 'xlsx', 'keynote', 'ppt', 'pptx']
        }
        Object.keys(typeMap).forEach((_type) => {
          const extensions = typeMap[_type]
          if (extensions.indexOf(extension) > -1) {
            type = _type
          }
        })
        return type
      },
      // 进度条
      progressStyle () {
        const progress = Math.floor(this.progress * 100)
        const style = `translateX(${Math.floor(progress - 100)}%)`
        return {
          progress: `${progress}%`,
          webkitTransform: style,
          mozTransform: style,
          msTransform: style,
          transform: style
        }
      },
      // 计算平均速度
      formatedAverageSpeed () {
        return `${Uploader.utils.formatSize(this.averageSpeed)} / s`
      },
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
      // 格式化剩余的时间
      formatedTimeRemaining () {
        const timeRemaining = this.timeRemaining
        const file = this.file
        if (timeRemaining === Number.POSITIVE_INFINITY || timeRemaining === 0) {
          return ''
        }
        let parsedTimeRemaining = secondsToStr(timeRemaining)
        const parseTimeRemaining = file.uploader.opts.parseTimeRemaining
        if (parseTimeRemaining) {
          parsedTimeRemaining = parseTimeRemaining(timeRemaining, parsedTimeRemaining)
        }
        return parsedTimeRemaining
      }
    },
    watch: {
      // 监听状态的变化，这个状态是个计算属性
      status (newStatus, oldStatus) {
        // 制作一个动效，如果新值和旧值都是uploading，那就给进度条的class属性置为 uploader-file-progressing
        // 如果不是uploading，去掉进度条的clss属性
        if (oldStatus && newStatus === 'uploading' && oldStatus !== 'uploading') {
          this.tid = setTimeout(() => {
            this.progressingClass = 'uploader-file-progressing'
          }, 200)
        } else {
          clearTimeout(this.tid)
          this.progressingClass = ''
        }
      }
    },
    methods: {
      //
      _actionCheck () {
        this.paused = this.file.paused
        this.typeError = this.file.typeError
        this.emptyError = this.file.emptyError
        this.error = this.file.error
        this.isUploading = this.file.isUploading()
      },
      // 暂停
      pause () {
        this.file.pause()
        this._actionCheck()
        this._fileProgress()
      },
      // 继续
      resume () {
        this.file.resume()
        this._actionCheck()
      },
      // 删除文件
      remove () {
        this.file.cancel()
      },
      // 重试
      retry () {
        this.file.retry()
        this._actionCheck()
      },
      processResponse (message) {
        let res = message
        try {
          res = JSON.parse(message)
        } catch (e) {
        }
        this.response = res
      },
      fileEventsHandler (event, args) {
        const rootFile = args[0]
        const file = args[1]
        const target = this.list ? rootFile : file
        if (this.file === target) {
          if (this.list && event === 'fileSuccess') {
            this.processResponse(args[2])
            return
          }
          this[`_${event}`].apply(this, args)
        }
      },
      _fileProgress () {
        this.progress = this.file.progress()
        this.averageSpeed = this.file.averageSpeed
        this.currentSpeed = this.file.currentSpeed
        this.timeRemaining = this.file.timeRemaining()
        this.uploadedSize = this.file.sizeUploaded()
        this._actionCheck()
      },
      _fileSuccess (rootFile, file, message) {
        if (rootFile) {
          this.processResponse(message)
        }
        this._fileProgress()
        this.error = false
        this.typeError = false
        this.emptyError = false
        this.isComplete = true
        this.isUploading = false
      },
      _fileComplete () {
        this._fileSuccess()
      },
      _fileError (rootFile, file, message) {
        this._fileProgress()
        this.processResponse(message)
        this.error = true
        this.typeError = true
        this.emptyError = true
        this.isComplete = false
        this.isUploading = false
      }
    },
    mounted () {
      const staticProps = ['paused', 'error', 'typeError','emptyError', 'averageSpeed', 'currentSpeed']
      const fnProps = [
        'isComplete',
        'isUploading',
        {
          key: 'size',
          fn: 'getSize'
        },
        {
          key: 'formatedSize',
          fn: 'getFormatSize'
        },
        {
          key: 'uploadedSize',
          fn: 'sizeUploaded'
        },
        'progress',
        'timeRemaining',
        {
          key: 'type',
          fn: 'getType'
        },
        {
          key: 'extension',
          fn: 'getExtension'
        }
      ]
      staticProps.forEach(prop => {
        this[prop] = this.file[prop]
      })
      fnProps.forEach((fnProp) => {
        if (typeof fnProp === 'string') {
          this[fnProp] = this.file[fnProp]()
        } else {
          this[fnProp.key] = this.file[fnProp.fn]()
        }
      })

      const handlers = this._handlers = {}
      const eventHandler = (event) => {
        handlers[event] = (...args) => {
          this.fileEventsHandler(event, args)
        }
        return handlers[event]
      }
      events.forEach((event) => {
        this.file.uploader.on(event, eventHandler(event))
      })
    },
    destroyed () {
      events.forEach((event) => {
        this.file.uploader.off(event, this._handlers[event])
      })
      this._handlers = null
    }
  }
</script>

<style>
  .uploader-file {
    position: relative;
    height: 49px;
    line-height: 49px;
    overflow: hidden;
    border-bottom: 1px solid #cdcdcd;
  }

  .uploader-file[status="waiting"] .uploader-file-pause,
  .uploader-file[status="uploading"] .uploader-file-pause {
    display: block;
  }

  .uploader-file[status="paused"] .uploader-file-resume {
    display: block;
  }

  .uploader-file[status="error"] .uploader-file-retry {
    display: block;
  }


  .uploader-file[status="success"] .uploader-file-remove {
    display: none;
  }

  .uploader-file[status="error"] .uploader-file-progress {
    background: #ffe0e0;
  }

  .uploader-file[status="typeError"] .uploader-file-retry {
    display: block;
  }

  .uploader-file[status="typeError"] .uploader-file-progress {
    background: #ffe0e0;
  }

  .uploader-file-progress {
    position: absolute;
    width: 100%;
    height: 100%;
    background: #e2eeff;
    transform: translateX(-100%);
  }

  .uploader-file-progressing {
    transition: all .4s linear;
  }

  .uploader-file-info {
    position: relative;
    z-index: 1;
    height: 100%;
    overflow: hidden;
  }

  .uploader-file-info:hover {
    background-color: rgba(240, 240, 240, 0.2);
  }

  .uploader-file-info i,
  .uploader-file-info em {
    font-style: normal;
  }

  .uploader-file-name,
  .uploader-file-size,
  .uploader-file-meta,
  .uploader-file-status,
  .uploader-file-actions {
    float: left;
    position: relative;
    height: 100%;
  }

  .uploader-file-name {
    width: 45%;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    text-indent: 14px;
  }

  .uploader-file-icon {
    width: 24px;
    height: 24px;
    display: inline-block;
    vertical-align: top;
    margin-top: 13px;
    margin-right: 8px;
  }

  .uploader-file-icon::before {
    content: "📃";
    display: block;
    height: 100%;
    font-size: 24px;
    line-height: 1;
    text-indent: 0;
  }

  .uploader-file-icon[icon="folder"]::before {
    content: "📂";
  }

  .uploader-file-icon[icon="image"]::before {
    content: "📊";
  }

  .uploader-file-icon[icon="video"]::before {
    content: "📹";
  }

  .uploader-file-icon[icon="audio"]::before {
    content: "🎵";
  }

  .uploader-file-icon[icon="document"]::before {
    content: "📋";
  }

  .uploader-file-size {
    width: 13%;
    text-indent: 10px;
  }

  .uploader-file-meta {
    width: 8%;
  }

  .uploader-file-status {
    width: 24%;
    text-indent: 20px;
  }

  .uploader-file-actions {
    width: 10%;
  }

  .uploader-file-actions > span {
    display: none;
    float: left;
    width: 16px;
    height: 16px;
    margin-top: 16px;
    margin-right: 10px;
    cursor: pointer;
    background: url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACgAAABkCAYAAAD0ZHJ6AAAAIGNIUk0AAHolAACAgwAA+f8AAIDpAAB1MAAA6mAAADqYAAAXb5JfxUYAAAAJcEhZcwAACxMAAAsTAQCanBgAAARkSURBVGje7ZnfS1NRHMAH4ptPkvQSuAdBkCxD8FUQJMEULUgzy1KyyPVQ4JMiiP4Bvg6EwUQQfMmwhwRDshwaKUjDVCgoSdDNHkzTJZ6+Z37Purve8+PeTb2TM/ggu+ew89l33x8H9BBCPG7GowXTJej3+wnDvEm0JuLC04+EYWftVAUv+fiCvDUdQR1BHUEdQR3BTIygvixoQS14XgTtthLVdpNWwXRLqvQ724LplFRtyrYF0yVpFLQrKRVMh6RZ0I6kkmCqklaCqpKZH0FX56Crq9jVfdDVk0RfFrSgFsxkQVmLcdKCVrKySCrryhPEyYShhzOcrFtG0EoilfHHk1CRU5rF6ZjNZhlVOW6RnMSVyyilKies4pO41diVy8wIujoHXV3FGdMHXTtJKLFYTLhZtq4vC1rwXApCZTIqgR6g1PBMCO9DL3bMMSqBHqDU8EyISDAHiGKvWwcCQG2KgjlAFCDAOhAAap0K5gKLphk8mqJgLrCIgoxRJ4J5wKpJ7gAoMkn5EBXBPGDVJHcAFJmkfIhQcAql1oBpTvTol9gG9pm4RHAKpdaAaU706JfYBvaZuJVgPQrt4sFlnOh5MC/p3lmJYD0K7eLBZZzoeTAv6d5ZnuAYHjpgEOnk5F0ufhG6v1ggOIaHDhhEOjl5l4tfhO4vthLcwAMrFNvLJO5vEwhu4IEViu1lEve3WQmyoihQFBzG/V0CQVYUBYqCw7i/SxTBcpsRbFeIYLnNCLZbCY5b5KAnxRwct8hBj9McZFVMW0ihRNBuFdMWUigRlFaxuQ9WWYjRMTiIe5z0wSoLMToGB3GPsA9aTZIJoB+nRgBnM1tzOkkmgH6cGgGczWzNpzqLx3n/aULJJgezeNw07oxQySbVywKjBOgFRnDs+VEsx8FlgVEC9AIjOPb8KJYjvSzoG7UW1IJaUAtqQS14toLNM5fN5APdwBJA8G83Pk/aK/rgzVvXzeQD3cASQPBvNz5P2ssTzAaGUIrHEO6zI5gNDKEUjyHcxxWkh4Ylcowwk1QQpIeGJXKMMJO0EgwqyjGCioJBJvDrxRMSuVOTJEXfbz1/bHwWtBL0yoQehK6RucgE+bGzanzulQh6E3IgQV+xpc8kcrfuSO7eTfJ3ZYmQw0Oy9azVKOk1C/bJ5D5F38YPeLfx0rjWJxHsS0SqsSYuxySjj5qO5Oj7xQWy2VBtFOwzCy6ryH3YfE3uh64Y1xckgstJPydEjkkeHv07Iy4Xaao15+KCWTBx6M/db+T9xivSErqaJDdzXI6yLRE8Vgg0coex/SPJvT0SbWu0KpZtbgSpCH3NRt7I5OxHkObc6heU+/M/J5vrpBFM5GBLqCQux14COXs5CNXK5OjPGm1tSMrJSOMNYQ4mVTGV/L6zTL7+DovkbFUxbSW0Wo05l8hJWsU+cRWfSh+Mt5Lb1ck/J1TvVsdDaR/MiEni+llsdZuZp62EViu+96bpNjNPWwmtVnzvFd5m9IVVC54x/wA7gNvqFG9vXQAAAABJRU5ErkJggg==") no-repeat 0 0;
  }

  .uploader-file-actions > span:hover {
    background-position-x: -21px;
  }

  .uploader-file-actions .uploader-file-pause {
    background-position-y: 0;
  }

  .uploader-file-actions .uploader-file-resume {
    background-position-y: -17px;
  }

  .uploader-file-actions .uploader-file-retry {
    background-position-y: -53px;
  }

  .uploader-file-actions .uploader-file-remove {
    display: block;
    background-position-y: -34px;
  }
</style>
