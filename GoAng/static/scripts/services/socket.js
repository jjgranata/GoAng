'use strict';

angular.module('GoAng')
    .factory('socket', ['socketFactory', function (socketFactory) {
        var socket = socketFactory();
        socket.forward('error');
        return socket;
    }]);