<template>
  <!-- 用于展示批量上传的所有文件列表 -->
  <div class="uploader-list">
    <slot :file-list="fileList">
      <ul>
        <li v-for="file in fileList" :key="file.id">
          <uploader-file :file="file" :list="true"></uploader-file>
        </li>
      </ul>
    </slot>
  </div>
</template>

<script>
  import { uploaderMixin } from '../common/mixins'
  import UploaderFile from './file.vue'

  const COMPONENT_NAME = 'uploader-list'

  export default {
    name: COMPONENT_NAME,
    mixins: [uploaderMixin],
    computed: {
      /* 他所展示的就是uploader的fileList */
      fileList () {
        return this.uploader.fileList
      }
    },
    components: {
      UploaderFile
    }
  }
</script>

<style>
  .uploader-list {
    position: relative;
  }
  .uploader-list > ul {
    list-style: none;
    margin: 0;
    padding: 0
  }
</style>
