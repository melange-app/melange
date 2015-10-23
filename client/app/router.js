var Components = require('./components');
var Routes = require('./routes');

var Router = Components.createStateful({
    stateName: "url",
    filterState: function(s) {
        return s.url;
    },
    selectPage: function(url) {
        var current = url.get('route');
        var route = Routes[current];

        // Use default route for the beginning, go to
        // notFound if we cannot load a route.
        if (current == undefined) {
            route = Routes.default;
        } else if (route == undefined) {
            route = Routes.notFound;
        }

        return route.page;
    },
    render: function() {
        var page = React.createElement(
            this.selectPage(this.state.url),
            {
                data: this.state.url.get('data'),
                route: this.state.url.get('route'),
            }
        );
        
        return (
            <div className="page">
                {page}
            </div>
        );
    }
});

module.exports = Router;
