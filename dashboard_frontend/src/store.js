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
      services: [],
    },
    requestSpanList: {
      items: [],
      total: 0,
    },
    serviceDownLogList: {
      items: [],
      total: 0,
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
    setRequestSpanList (state, payload) {
      if (!payload) {
        return
      }
      state.requestSpanList = payload
    },
    setServiceDownLogList (state, payload) {
      if (!payload) {
        return
      }
      state.serviceDownLogList = payload
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
    loadRequestSpanList ({ commit }, { offset, limit }) {
      return callApi('/api/list_request_span', 'POST', {
        offset: offset || 0,
        limit: limit || 20,
      }).then(res => {
        console.log('request spans', res)
        commit('setRequestSpanList', res)
        return res
      })
    },
    loadServiceDownLogList ({ commit }, { offset, limit }) {
      return callApi('/api/list_service_down_logs', 'POST', {
        offset: offset || 0,
        limit: limit || 20,
      }).then(res => {
        console.log('service down logs', res)
        commit('setServiceDownLogList', res)
        return res
      })
    },
    queryServiceHealthByUrl ({ commit }, serviceUrl) {
      (() => {})(commit)
      return callApi('/api/query_service_health', 'POST', {
        url: serviceUrl,
      })
    },
  },
  getters: {
    upstreamList (state) {
      return state.statistics.upstreamServices
    },
    serviceList (state) {
      return state.statistics.services
    },
    requestList (state) {
      return state.requestSpanList
    },
    serviceDownLogList (state) {
      return state.serviceDownLogList
    },
  },
})
