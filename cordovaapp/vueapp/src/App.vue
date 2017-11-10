<template>
  <div id="app">
    <header>
        <icon id="selmuseum" name="bars" scale=1.5 @click.native="selMuseum"></icon>
        <span>Boston Fine Art</span>
        <icon id="takephoto" name="camera" scale=1.5 @click.native="takePhoto"></icon>
    </header>
    <div id="contentview">
      <router-view/>
    </div>
    <nav-footer></nav-footer>
  </div>
</template>

<script>
import NavFooter from '@/components/NavFooter'
export default {
  name: 'app',
  components: {
    "nav-footer": NavFooter
  },
  methods: {
    takePhoto() {
      this.$router.push({path:"predict", query: { action: 'takephoto', random: Math.random(), time: Date.now() }})
    },

    selMuseum() {
      console.log('selMuseum');
    }
  }
}
</script>