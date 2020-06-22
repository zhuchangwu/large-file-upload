import Uploader from './components/uploader.vue'
import UploaderBtn from './components/btn.vue'
import UploaderDrop from './components/drop.vue'
import UploaderUnsupport from './components/unsupport.vue'
import UploaderList from './components/list.vue'
import UploaderFiles from './components/files.vue'
import UploaderFile from './components/file.vue'

const uploader = {
  version: /* eslint-disable no-undef */ __VERSION__,
  install,
  Uploader,
  UploaderBtn,
  UploaderDrop,
  UploaderUnsupport,
  UploaderList,
  UploaderFiles,
  UploaderFile
}

if (typeof window !== 'undefined' && window.Vue) {
  window.Vue.use(install)
}

export default uploader

function install (Vue) {
  if (install.installed) {
    return
  }
  // 直接在vue上注册全局组件，所以我们能直接使用这些组件
  Vue.component(Uploader.name, Uploader)
  Vue.component(UploaderBtn.name, UploaderBtn)
  Vue.component(UploaderDrop.name, UploaderDrop)
  Vue.component(UploaderUnsupport.name, UploaderUnsupport)
  Vue.component(UploaderList.name, UploaderList)
  Vue.component(UploaderFiles.name, UploaderFiles)
  Vue.component(UploaderFile.name, UploaderFile)
}
