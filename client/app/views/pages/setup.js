var S = require("../../store");
var AsyncButton = require("../../components").AsyncButton;

var ConfirmMelange = React.createClass({
    gotoHome: function() {
        var R = require("../../routes");
        S.goto(R.home);
    },
    performRegistration: function() {
        return Q.delay(1000);
    },
    render: function() {
        return (
            <div className="container setup-white">
                <h1>Review Details <small>Step 3 of 3</small></h1>

                <p>
                    Please confirm your details before performing the
                    registration.
                </p>

                <h3>Server Selection</h3>
                <p>Melange Development Server</p>

                <h3>Username Selection</h3>
                <p>hleath</p>

                <AsyncButton onClick={ this.performRegistration } onDone={ this.gotoHome }>
                    Register
                </AsyncButton>
            </div>
        );
    }
});

var NameSelect = React.createClass({
    getInitialState: function() {
        return {
            verified: false,
        };
    },
    gotoReview: function() {
        var R = require("../../routes");
        S.goto(R.setup, {
            name: "newConfirmation",
        });
    },
    verifyName: function() {
        return Q.delay(1000);
    },
    render: function() {
        return (
            <div className="container setup-white">
                <h1>Select a Name <small>Step 2 of 3</small></h1>
                
                <p>
                    The name that you select here will be equivalent
                    to your address on the Melange system. Your
                    friends can send you messages with it.
                </p>
                
                <p>
                    Melange uses the Namecoin Blockchain to keep your
                    name information secure. However, registering a
                    name may take up to a couple of hours. Our system
                    will automatically perform the final registration
                    steps and serve your name before the registration
                    is complete. At no point will the server have your
                    private Namecoin keys.
                </p>
                
                <p>
                    During the initial Melange offering, names will be
                    provided freely.
                </p>

                <form>
                    <div className="form-group">
                        <label for="name">Your Username</label>
                        <input id="name" placeholder="New Username..." className="form-control"/>
                    </div>
                    <div className="pull-right">
                        <AsyncButton onClick={ this.verifyName } onDone={ this.gotoReview }>
                            Check Availability
                        </AsyncButton>
                    </div>
                </form>
            </div>
        );
    }
});

var ServerSelect = React.createClass({
    gotoName: function() {
        var R = require("../../routes");
        S.goto(R.setup, {
            name: "newName",
        })
    },
    render: function() {
        return (
            <div className="container setup-white">
                <h1>Select a Server <small>Step 1 of 3</small></h1>
                <p>
                    In order to use Melange, you must select which
                    server will host your data in the cloud. Melange
                    prevents these services from reading any of your
                    data, but you may wish to select the server based
                    on cost or availability.
                </p>
                <div className="row">
                    <div className="col-xs-6">
                        <div className="form-group">
                            <input className="form-control" type="text" placeholder="Filter..."/>
                        </div>
                        <div className="list-group">
                            <a className="list-group-item">
                                Hello, world.
                            </a>
                            <a className="list-group-item">
                                Hello, world.
                            </a>
                        </div>
                    </div>
                    <div className="col-xs-6">
                        <div className="panel">
                            <div className="panel-body">
                                <h3>Development Server</h3>

                                <p>
                                    This server is hosted by Melange.
                                </p>

                                <hr/>

                                <div onClick={ this.gotoName } className="btn btn-block btn-primary">
                                    Select this Server &raquo;
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
});

var Link = React.createClass({
    render: function() {
        return (
            <h1>Select a Linky</h1>
        );
    }
});

var Welcome = React.createClass({
    gotoSignup: function() {
        var R = require("../../routes");
        S.goto(R.setup, {
            name: "newServer",
        });
    },
    gotoLink: function() {
        var R = require("../../routes");
        S.goto(R.setup, {
            name: "link",
        })
    },
    render: function() {
        return (
            <div className="setup">
                <p className="lead">Welcome to</p>
                
                <br/>
                
                <div className="row">
                    <div className="col-xs-3">
                        <img className="img-responsive" src="/img/icon.png"/>
                    </div>
                    <div className="col-xs-8">
                        <h1 className="style">Melange</h1>
                        <h3>Developer Preview</h3>
                    </div>
                </div>
                
                <br/>
                
                <div className="row">
                    <div className="col-xs-6">
                        <p className="lead">
                            The secure social network that gives you
                            back control of your data.
                        </p>

                        <div className="row text-center">
                            <div className="col-xs-6">
                                <h1><i className="fa fa-cloud-download"></i></h1>
                                <p>Install Apps</p>
                            </div>
                            <div className="col-xs-6">
                                <h1><i className="fa fa-users"></i></h1>
                                <p>Find Friends</p>
                            </div>
                        </div>
                    </div>
                    <div className="col-xs-6">
                        <SetupButton onClick={ this.gotoSignup } text="Sign Up"/>
                        <SetupButton onClick={ this.gotoLink } text="Link Existing Account" color="pink"/>
                    </div>
                </div>
            </div>
        )
    },
});

var SetupButton = React.createClass({
    render: function() {
        var style = {};
        if (this.props.color != undefined) {
            style["borderBottomColor"] = this.props.color;
        }
        
        return (
            <div onClick={ this.props.onClick } className="button" style={style}>
                <p>{ this.props.text } &rarr;</p>
            </div>
        );
    },
});

// Setup is a router that handles the different pages inside of the setup module.
var Setup = React.createClass({
    render: function() {
        if (this.props.data === undefined) {
            return (
                <Welcome {...this.props}/>
            )
        }
        

        var route = routes[this.props.data.get('name')];
        if (route == undefined) {
            console.log("Unable to find the correct settings route", this.props.data.get('name'));
        }
        
        return React.createElement(route, this.props);
    }
});

var routes = {
    newServer: ServerSelect,
    newName: NameSelect,
    newConfirmation: ConfirmMelange,
    link: Link,
}

console.log(routes);

module.exports = Setup;
