'use strict'

var test = require('tape').test
var memory = require('./')
var abs = require('./abstract')

abs({
  test: test,
  persistence: memory
})
