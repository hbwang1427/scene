<template>
	<div id="profile">
		hello @@

		<p><a @click.prevent="signOut">sign out</a></p>
	</div>
</template>

<script>
export default {
	name:"Profile",
	data: function(){
		return {}
	},
	methods: {
		signOut() {
			localStorage.removeItem("authToken");
			this.$router.push({path:'predict'});
		}
	}
}
</script>