<template>
	<div id="login">	
		<div class="wrap">
			<div class="row">
				<svg class="icon name svg-icon" viewBox="0 0 20 20">
		            <path d="M0,20 a10,8 0 0,1 20,0z M10,0 a4,4 0 0,1 0,8 a4,4 0 0,1 0,-8"></path>
		        </svg>
				
				<input v-model="name" type="text" class="input name" placeholder="Username"/>
			</div>

			<div class="row">
				<svg class="icon pass svg-icon" viewBox="0 0 20 20">
	            	<path d="M0,20 20,20 20,8 0,8z M10,13 10,16z M4,8 a6,8 0 0,1 12,0"></path>
	          	</svg>
				
				<input v-model="password" type="password" class="input pass" placeholder="Password"/>
			</div>

			<div class="row" v-if="error">
				{{error}}
			</div>

			<button type="button" class="submit" @click.prevent="login">Sign in</button>

			<p class="signup">Don't have an account? &nbsp;<a>Sign up</a></p>
		</div>
	</div>
</template>


<script>
import router from '../router';

export default {
	name: 'Login',
	data() {
		return {
			error: '',
			name: '123',
			password: ''
		}
	},
	methods: {
		login: function() {
			if (this.name == "123") {
				localStorage.setItem("authToken", "temptoken");
				this.error = "";
				var redirect = this.$route.query.redirect;
				console.log("you are login, forward to:");
				console.log(this.$route);
				router.replace(redirect);
			} else {
				this.error = "invalid username or password";
			}
		},

		cancel: function() {
			router.replace('/');
		}
	}
}
</script>