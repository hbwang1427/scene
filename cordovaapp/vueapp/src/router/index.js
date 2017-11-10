import Vue from 'vue'
import Router from 'vue-router'
import Main from '@/components/Main'
import Login from '@/components/Login'
import Predict from '@/components/Predict'
import SearchMuseum from '@/components/SearchMuseum'
import Profile from '@/components/Profile'
import NavFooter from '@/components/NavFooter'
import store from '../store'
Vue.use(Router)

const router = new Router({
  routes: [
    {
      path: '/',
      component: Main,

      children: [
      	{
      		path: '',
      		component: Predict
      	},
      	{
      		path: 'predict',
      		component: Predict,
      	},
      	{
      		path: 'profile',
      		component: Profile,
      		meta: { requiresAuth: true }
      	},
      	{
      		path: 'search',
      		component: SearchMuseum
      	}
      ]
    },
    {
      path: '/login',
      name: 'Login',
      component: Login
    }
  ]
});

router.beforeEach((to, from, next) => {
	//console.log(from.fullPath+"->"+to.fullPath);
  var authToken = localStorage.getItem("authToken");
	if (to.matched.some(record => record.meta.requiresAuth) && !authToken) {
		//console.log("redirect to /login");
    store.dispatch('setpage', '/login')
		next({
	        path: '/login',
	        query: { redirect: to.fullPath }
	      })
	} else {
    store.dispatch('setpage', to.fullPath);
		next();
	}
});


export default router