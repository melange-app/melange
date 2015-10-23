var S = require("../../store");

var RecentLink = React.createClass({
    handleClick: function() {
        this.props.onClick();
    },
    render: function() {
        return (
            <div className="recent"><div onClick={ this.handleClick } className="inner">
                <div className="icon" style={{ "backgroundImage": "url('/img/icon.png')" }}></div>
                <h1>Chat with Joey B</h1>
                <p>September 28, 2015</p>
            </div></div>
        );
    },
});

var Home = React.createClass({
    gotoSettings: function(e) {
        var Routes = require('../../routes');
        
        S.goto(Routes.settings);
        e.preventDefault();
    },
    gotoRecent: function() {
        S.goto(Routes.recents)
    },
    gotoApps: function(e) {
        S.dispatch(S.actions.views.menu);
        e.preventDefault()
    },
    render: function() {
        var recents = [
            <RecentLink onClick={ this.gotoRecent }/>,
            <RecentLink onClick={ this.gotoRecent }/>,
            <RecentLink onClick={ this.gotoRecent }/>,
            <RecentLink onClick={ this.gotoRecent }/>,
        ];
        
        return (
            <div className="home">
                <div className="home-inner">
                    <div className="search">
                        <input type="text" placeholder="Search Melange..."/>
                    </div>

                    <div className="actions">
                        <div className="pull-right">
                            <a onClick={ this.gotoSettings } href="">
                                <i className="fa fa-fw fa-gears"></i>
                                Settings
                            </a>
                        </div>
                        <a href="">
                            <i className="fa fa-fw fa-pencil"></i>
                            Post
                        </a>
                    </div>

                    <div className="alert alert-danger">
                        You need to renew your name. Please do so now.
                    </div>

                    <div className="recents">
                        { recents }
                    </div>

                    <hr/>

                    <div className="apps">
                        <h4><a onClick={ this.gotoApps } href="">
                            <i className="fa fa-fw fa-list-alt"></i>
                            Apps
                        </a></h4>
                    </div>

                    <hr/>

                    <div className="people">
                        <h4><a href="">
                            <i className="fa fa-fw fa-users"></i>
                            People
                        </a></h4>
                    </div>
                </div>
            </div>
        );
    }
});

module.exports = Home
