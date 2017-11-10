<template>
<footer>
	<ul>
		<li>
			<icon name="question-circle-o" scale=1.5 @click.native="topage('/predict')" v-bind:class="{active : ispage('/predict')}"></icon>
		</li>
		<li>
			<icon name="search" scale=1.5 @click.native="topage('/search')" v-bind:class="{active : ispage('/search')}"></icon>
		</li>
		<li>
			<icon name="user-o" scale=1.5 @click.native="topage('/profile')" v-bind:class="{active : ispage('/profile')}"></icon>
		</li>
	</ul>
</footer>
</template>

<script>
import { mapState } from 'vuex'
export default {
	name:"navFooter",
	data() {
		return {
		}
	},
	methods: {
		topage: function(page) {
			this.$router.push({path: page});
		},
		ispage(page) {
			return this.$store.state.page.indexOf(page) >= 0;
		}
	},
	computed: {
		page() {
			return this.$store.state.page;
		}
	}
}
</script>
