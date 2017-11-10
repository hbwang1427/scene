// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
import router from './router'

require('./bootstrap');

//import all icons if you don't care about bundle size
import 'vue-awesome/icons'
import Icon from 'vue-awesome/components/Icon'
Vue.component('icon', Icon)

//import global css
require("./assets/app.scss");

//promise polyfill
require('es6-promise').polyfill();

import store from './store'

Vue.config.productionTip = false


if (process.env.NODE_ENV === "development") {
	//for browser
	new Vue({
	  el: '#app',
	  store,
	  router,
	  template: '<App/>',
	  components: { App }
	})
} else {
	//for android
	document.addEventListener('deviceready', function() {
		console.log(device.cordova);
		// if (AndroidFullScreen) {
		// 	AndroidFullScreen.showUnderSystemUI(
		// 		()=>{console.log("showUnderSystemUI success")}, 
		// 		()=>{});
		// } 


		navigator.geolocation.getCurrentPosition(function(position) {
			store.dispatch('updategeo', position);
	    },  function (error) {
	    	console.log("get geoposition error:" + error.message);
	    }, { timeout: 3000 });

		new Vue({
		  el: '#app',
		  store,
		  router,
		  template: '<App/>',
		  components: { App }
		});
	   
	    //window.navigator.splashscreen.hide();
	}, false);
}