'use strict';

/* Filters */

var melangeFilters = angular.module('melangeFilters', [])

melangeFilters.filter('unsafe', ['$sce', function($sce) {
    return function(val) {
        return $sce.trustAsHtml(val);
    };
}]);

melangeFilters.filter('unsafeUrl', ['$sce', function($sce) {
    return function(val) {
        return $sce.trustAsResourceUrl(val);
    };
}]);
