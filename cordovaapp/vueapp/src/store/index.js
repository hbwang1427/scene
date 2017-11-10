import Vue from 'vue'
import Vuex from 'vuex'
import * as types from './mutation-types'

Vue.use(Vuex)
const debug = process.env.NODE_ENV !== 'production'

//initial state
const state = {
	page: "",
	geo: {
		latitude:0,
		longitude:0
	}
}

// getters
const getters = {
  latitude: state => state.geo.latitude,
  longitude: state => state.geo.longitude,
  page: state => state.page,
}


// actions
const actions = {
	updategeo ({ commit, state }, position) {
		commit(types.UPDATE_GEOLOCATION, {
			latitude:position.coords.latitude, 
			longitude:position.coords.longitude
		});
	},

	setpage({commit, state}, page) {
		commit(types.SET_PAGE, {page})
	}
}


// mutations
const mutations = {
  [types.UPDATE_GEOLOCATION] (state, { latitude, longitude }) {
  	state.geo = {latitude, longitude}
  },

  [types.SET_PAGE] (state, { page }) {
  	state.page = page
  }
}

export default new Vuex.Store({
  state,
  actions,
  getters,
  mutations
})