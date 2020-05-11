import axios from 'axios'
import Vue from 'vue'

if (process.env.NODE_ENV === 'development') {
    // 开发环境
    axios.defaults.baseURL = 'http://127.0.0.1:5000'
} else {
    // 真实环境
    axios.defaults.baseURL = 'http://127.0.0.1:5000'
}
axios.defaults.withCredentials = false
axios.defaults.headers = { 'Access-Control-Allow-Origin': '*' }
// axios.defaults.mode = 'no-cors'

const instance = axios.create({
    baseURL: axios.defaults.baseURL,
    headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PATCH, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Origin, Content-Type, X-Auth-Token, Access-Control-Allow-Origin, Access-Control-Allow-Methods, Content-Length',
        // 'Sec-Fetch-Dest': 'text',
        // 'Sec-Fetch-Mode': 'no-cors',
        // 'Sec-Fetch-Site': 'site',
    },
    withCredentials: false,
    // mode: 'no-cors',
})

instance.interceptors.request.use(function (config) {
    // 在发送请求之前做些什么
    return config
}, function (error) {
    // 对请求错误做些什么
    return Promise.reject(error)
})

Vue.prototype.$axios = instance

export default instance
