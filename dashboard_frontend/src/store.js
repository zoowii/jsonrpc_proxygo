import Vue from 'vue'
import Vuex from 'vuex'
import axios from './axios/request'

Vue.use(Vuex)

// const axios = Vue.prototype.$axios

const ignoreUnused = (...args) => {
  (() => { })(args)
}
ignoreUnused()

const callApi = (url, method, data) => {
  return axios({
    method: method,
    url: url,
    // mode: 'no-cors',
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Content-Type': 'application/json',
    },
    // withCredentials: false,
    data: data,
  }).then(response => {
    // console.log('response', response)
    if (response.status !== 200) {
      throw new Error(response.body)
    }
    const data = response.data
    if (data.error) {
      throw new Error(data.error)
    }
    return data
  }).then(result => {
    console.log('result', result)
    return result
  })
    .catch(err => {
      console.log('error', err)
      throw err
    })
}

export default new Vuex.Store({
  state: {
    barColor: 'rgba(0, 0, 0, .8), rgba(0, 0, 0, .8)',
    barImage: 'https://demos.creative-tim.com/material-dashboard/assets/img/sidebar-1.jpg',
    drawer: null,
    statistics: {
      globalRpcCallCount: 0,
      hourlyRpcCallCount: 0,
      globalStat: {},
      hourlyStat: {},
      upstreamServices: [],
    },
  },
  mutations: {
    SET_BAR_IMAGE (state, payload) {
      state.barImage = payload
    },
    SET_DRAWER (state, payload) {
      state.drawer = payload
    },
    setStatistics (state, payload) {
      if (!payload) {
        return
      }
      state.statistics = payload
    },
  },
  actions: {
    loadStatistic ({ commit }) {
      return callApi('/api/statistic', 'GET', {})
        .then(res => {
          console.log('statistics', res)
          commit('setStatistics', res)
        })
    },
  },
})
