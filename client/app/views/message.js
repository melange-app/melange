var S = require('../store');
var Viewer = require('./plugin').Viewer;

var Message = React.createClass({
    gotoProfile: function(e) {
        e.preventDefault();
        
        var Routes = require('../routes');

        // close the menu and goto
        S.dispatch(S.actions.views.newsfeed);
        S.goto(
            Routes.profile,
            {
                id: this.props.message.getIn(["from", "alias"]),
            }
        );
    },
    render: function() {
        var name = this.props.message.getIn(['from', 'name']);
        if (name == "") {
            name = this.props.message.getIn(['from']);
        }

        // TODO: Implement image proxy on server to prevent XSS attacks.
        var avatar = this.props.message.getIn(['from', 'avatar']);
        if (avatar == "") {
            avatar = "/img/icon.png";
        }
        var avatarStyle = {
            "backgroundImage": "url(\"" + encodeURI(avatar) + "\")"
        }

        var time = this.props.message.get("_parsedDate");
        var timeString = time.fromNow();
        if (time.isBefore(moment().subtract(1, 'days'))) {
            timeString = time.calendar();
        }
        
        return (
            <div className="story">
                <div onClick={ this.gotoProfile }
                     className="avatar"
                     style={ avatarStyle }></div>

                <h3><a onClick={ this.gotoProfile } href="">{ name }</a></h3>
                <p className="muted" title={ time.format('MMMM Do YYYY, h:mm a') }>
                    { timeString }
                </p>

                <Viewer message={ this.props.message }/>

                <div className="actions">
                    <a href="">Like</a> &middot;
                    <a href="">Comment</a>
                </div>
            </div>
        );
    } 
});

module.exports = Message;
