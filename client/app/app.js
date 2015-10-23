var Status = require('./views/status');
var NewsFeed = require('./views/newsfeed');
var Menu = require('./views/menu');
var Toolbar = require('./views/toolbar');

var Router = require('./router');

var Components = require('./components');

window.images = {
    hunter: "https://scontent-iad3-1.xx.fbcdn.net/hphotos-xpf1/v/t1.0-9/11755276_1598154330449890_3860781029961901461_n.jpg?oh=bf9cfab789484f74fbfe92fb8a3e5d11&oe=56A47885",
    cover: "http://www.smittenblogdesigns.com/wp-content/uploads/2014/01/passion1.png",
}

window.getBackground = function(name) {
    var img = images[name];
    if (name == undefined) {
        img = "/img/icon.png"
    }
    
    return {
        "backgroundImage": "url('" + img + "')"
    }
}

// Create Redux Store
var S = require("./store");

var Backend = require("./services/backend");

// Export store globally for debugging... (probably a bad idea!)
window.melange = {
    store: S.store,
    backend: Backend,
}

var Body = Components.createStateful({
    stateName: "f",
    filterState: function(s) {
        return s.views;
    },
    render: function() {
        var menuOpen = {
            "body": true,
            "menu-open": this.state.f.get('menu'),
            "newsfeed-open": this.state.f.get('newsfeed'),
        }
        
        return (
            <div className={ Components.classSet(menuOpen) }>
                <NewsFeed/>
                <Menu/>
                <Router/>
            </div>
        )
    },
});

var Melange = React.createClass({
    render: function() {
        return (
            <div className="react">
                <Toolbar/>
                <Body/>
                <Status/>
            </div>
        )
    },
});

React.render(<Melange/>, document.getElementById('melange'), function() {
    console.log("React application started.");
    
    var loader = document.getElementById("loader");
    loader.style.display = "none";

    Backend.realtime.connect();
});
