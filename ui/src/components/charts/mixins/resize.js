function debounce(func, wait, immediate) {
  let timeout, args, context, timestamp, result

  const later = function() {
    // 据上一次触发时间间隔
    const last = +new Date() - timestamp

    // 上次被包装函数被调用时间间隔 last 小于设定时间间隔 wait
    if (last < wait && last > 0) {
      timeout = setTimeout(later, wait - last)
    } else {
      timeout = null
      // 如果设定为immediate===true，因为开始边界已经调用过了此处无需调用
      if (!immediate) {
        result = func.apply(context, args)
        if (!timeout) context = args = null
      }
    }
  }

  return function(...args) {
    context = this
    timestamp = +new Date()
    const callNow = immediate && !timeout
    // 如果延时不存在，重新设定延时
    if (!timeout) timeout = setTimeout(later, wait)
    if (callNow) {
      result = func.apply(context, args)
      context = args = null
    }

    return result
  }
}

export default {
  data() {
    return {
      $_sidebarElm: null,
      $_resizeHandler: null
    }
  },
  mounted() {
    this.initListener()
  },
  activated() {
    if (!this.$_resizeHandler) {
      // avoid duplication init
      this.initListener()
    }

    // when keep-alive chart activated, auto resize
    this.resize()
  },
  beforeDestroy() {
    this.destroyListener()
  },
  deactivated() {
    this.destroyListener()
  },
  methods: {
    // use $_ for mixins properties
    // https://vuejs.org/v2/style-guide/index.html#Private-property-names-essential
    $_sidebarResizeHandler(e) {
      if (e.propertyName === 'width') {
        this.$_resizeHandler()
      }
    },
    initListener() {
      this.$_resizeHandler = debounce(() => {
        this.resize()
      }, 100)
      window.addEventListener('resize', this.$_resizeHandler)

      this.$_sidebarElm = document.getElementsByClassName(
        'sidebar-container'
      )[0]
      this.$_sidebarElm &&
        this.$_sidebarElm.addEventListener(
          'transitionend',
          this.$_sidebarResizeHandler
        )
    },
    destroyListener() {
      window.removeEventListener('resize', this.$_resizeHandler)
      this.$_resizeHandler = null

      this.$_sidebarElm &&
        this.$_sidebarElm.removeEventListener(
          'transitionend',
          this.$_sidebarResizeHandler
        )
    },
    resize() {
      const { Highcharts } = this
      Highcharts && Highcharts.resize()
    }
  }
}
