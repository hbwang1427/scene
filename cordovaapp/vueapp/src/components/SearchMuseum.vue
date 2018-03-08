<template>
	<div id="search">
		<p>search museum in this page.</p>
		<p>latitude: {{latitude}}</p>
		<p>longitude: {{longitude}}</p>

		<p v-if="myposition">

		</p>
		<p v-if="weather">
			description: {{weather.description}}
			<ul>
				<li>temp: {{Math.floor((weather.temp - 32)/1.8)}} ℃</li>
				<li>wind: {{weather.wind}}</li>
				<li>humidity: {{weather.humidity}}</li>
				<li>visibility: {{weather.visibility}}</li>
				<li>sunrise: {{weather.sunrise}}</li>
				<li>sunset: {{weather.sunset}}</li>
			</ul>
		</p>
	</div>
</template>

<script>
export default {
	name:"SearchMuseum",
	data: function(){
		return {
			myposition:""
		}
	},
	computed: {
		latitude() {
			return this.$store.state.geo.latitude;
		},

		longitude() {
			return this.$store.state.geo.longitude;
		},

		weather() {
			return this.$store.state.weather;
		}
	}
}
</script>