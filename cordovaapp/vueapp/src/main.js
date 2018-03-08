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
var geoWatchId = -1;

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

	function setupGeoLocationWatcher() {
		if (geoWatchId == -1) {
			geoWatchId = navigator.geolocation.watchPosition(function(position){
				console.log(position);
				var np = eviltransform.bd2gcj(position.coords.latitude, position.coords.longitude); // BD-09 -> GCJ-02
				position.coords.latitude = np.lat;
				position.coords.longitude = np.lng;
				store.dispatch('updategeo', position);
				getWeather(position.coords.latitude, position.coords.longitude);
			}, function(error){
				console.log("get geoposition error:" + error.message);
				alert("get geoposition error:" + error.message);
			}, {timeout: 30000, enableHighAccuracy: true, maximumAge: 75000});
		}
	}

	function getWeather(latitude, longitude) {
		var OpenWeatherAppKey = "f582f8a21b3b8b3f0f20a8a01789f217";
	    var queryString =
	      'http://api.openweathermap.org/data/2.5/weather?lat='
	      + latitude + '&lon=' + longitude + '&appid=' + OpenWeatherAppKey + '&units=imperial';
	    axios.get(queryString).then((response)=>{
	    	if (response.status == 200) {
	    		var results = response.data;
		    	if (results.weather.length) {
	            	store.dispatch("updateweather", {
	            		description: results.name,
	            		temp: results.main.temp,
	            		wind: results.wind.speed,
	            		humidity: results.main.humidity,
	            		visibility: results.weather[0].main,
	            		sunrise: new Date(results.sys.sunrise).toLocaleTimeString(),
	            		sunset: new Date(results.sys.sunset).toLocaleTimeString()
	            	});
		        }
	    	}
	    }).catch((error)=>{
	    	alert("get weather error:" + error);
	    });
	}

	//for android
	document.addEventListener('deviceready', function() {
		console.log(device.cordova);
		// if (AndroidFullScreen) {
		// 	AndroidFullScreen.showUnderSystemUI(
		// 		()=>{console.log("showUnderSystemUI success")}, 
		// 		()=>{});
		// } 

		new Vue({
		  el: '#app',
		  store,
		  router,
		  template: '<App/>',
		  components: { App }
		});
	   
	   	setupGeoLocationWatcher();
	    //window.navigator.splashscreen.hide();
	}, false);


	document.addEventListener("resume", function(){
		console.log("app resumed");
		setupGeoLocationWatcher();
	}, false);


	document.addEventListener("pause", function(){
		console.log("app paused");
		navigator.geolocation.clearWatch(geoWatchId);
		geoWatchId = -1;
	}, false);
}