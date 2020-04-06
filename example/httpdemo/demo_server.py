# coding: utf8

from flask import Flask
from flask_jsonrpc import JSONRPC

app = Flask(__name__)
jsonrpc = JSONRPC(app, '/api', enable_web_browsable_api=True)

@jsonrpc.method('App.index')
def index():
    return 'Welcome to Flask JSON-RPC'

@jsonrpc.method('hello')
def hello(name):
    return 'Hello, %s, this is response from server' % name

@jsonrpc.method("error")
def dummyError():
    raise Exception('dummy error response')

if __name__ == '__main__':
    app.run(port=3000, debug=True)
