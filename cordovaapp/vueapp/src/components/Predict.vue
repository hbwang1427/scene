<template>
	<div id="predict">
		<div id="loading" v-if="isRequesting">
			<icon  name="refresh" spin ></icon> computing...
		</div>

		<div id="detail">
			<img id="srcphoto" v-if="photourl && text.length==0" :src="photourl"/>
			<p v-if="text">{{text}}</sp>
			<p v-if="audioUrl"><audio id="paudio" controls="controls" :src="audioUrl"></audio></p>
			<p v-if="videoUrl">{{videoUrl}}</p>
		</div>

		<ul v-if="results" id="results">
			<li v-for="(item, index) in results">
				<img :src="item.image_url"  @click="displayText(index)"/>
			</li>
		</ul>
	</div>
</template>

<script>
export default {
	name:"Predict",
	data() {
		return {
			text: "",
			audioUrl:"",
			videoUrl:"",
			isRequesting: false,
			error:"",
			photourl:"",
			results:[
				
			]
		}
	},

	created () {
		console.log("predict created");
    	axios.post("http://100.0.245.19:8081/predict", qs.stringify({"image": this.photourl, "limits": 0})).then(response => {
			if (response.data.error) {
				this.error = response.data.error;
			} else {
				this.results = response.data.results;
			}
			this.isRequesting = false;
		  }).catch(function(error) {
		  	this.isRequesting = false;
		  	alert(error);
		  });
  	},

  	beforeRouteUpdate (to, from, next) {
  		console.log("before route update:" + to.fullPath);
  		var i = to.fullPath.indexOf("?");
  		if (i > 0) {
  			var query = _.chain(to.fullPath.substr(i+1))
  				.replace('?', '') 
			    .split('&') 
			    .map(_.ary(_.partial(_.split, _, '='), 1))
			    .fromPairs()
			    .value();
			if (query['action'] === 'takephoto') {
				console.log("launch camera");
				this.takePhoto();
			}
  		}
  		next();
  	},

	methods: {
		takePhoto: function() {
			this.photourl = "";
			navigator.camera.getPicture(this.onTakePhotoSuccess, this.onTakePhotoFail, {  
		       quality: 50,  
		       destinationType: Camera.DestinationType.DATA_URL,  
		       encodingType: Camera.EncodingType.JPEG,  
		       sourceType: Camera.PictureSourceType.CAMERA
		   });
		},

		onTakePhotoSuccess: function(imageData) {
			this.photourl = "data:image/jpeg;base64," + imageData;
			this.isRequesting = true;
			this.text = '';
			axios.post("http://100.0.245.19:8081/predict", qs.stringify({"image": this.photourl, "limits": 0})).then(response => {
				console.log('aitourscene:'+response.data);
				if (response.data.error) {
					this.error = response.data.error;
				} else {
					this.results = response.data.results;
				}
				this.isRequesting = false;
			  }).catch(function(error) {
			  	this.isRequesting = false;
			  	alert(error);
			  });
		},

		onTakePhotoFail: function(message) {
			alert('Failed because: ' + message);
		},

		displayText(index) {
			if (this.results.length > index) {
				this.text = this.results[index].text;
				this.audioUrl = this.results[index].audio_url;
				this.videoUrl = this.results[index].video_url;
			}
		}
	}
}
</script>
