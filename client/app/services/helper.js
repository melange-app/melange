(function() {
  var melangeServices = angular.module('melangeServices');
  melangeServices.factory('mlgHelper', ['$rootScope', function($rootScope) {
    return {
      promise: function(result, func) {
            if (!result || !func) return result;

            var then = function(promise) {
                //see if they sent a resource
                if ('$promise' in promise) {
                    promise.$promise.then(update);
                }
                //see if they sent a promise directly
                else if ('then' in promise) {
                    promise.then(update);
                }
            };

            var update = function(data) {
                if ($.isArray(result)) {
                    //clear result list
                    result.length = 0;
                    //populate result list with data
                    $.each(data, function(i, item) {
                        result.push(item);
                    });
                } else {
                    //clear result object
                    for (var prop in result) {
                        if (prop !== 'load') delete result[prop];
                    }
                    //deep populate result object from data
                    $.extend(true, result, data);
                }
            };

            //see if they sent a function that returns a promise, or a promise itself
            if ($.isFunction(func)) {
                // create load event for reuse
                result.load = function() {
                    then(func());
                };
                result.load();
            } else {
                then(func);
            }

            return result;
        },
      }
  }]);
})()
