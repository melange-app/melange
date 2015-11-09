(function() {
    var styles = function() {
        var links = document.getElementsByTagName('link');
        for (var i = 0; i<links.length; i++) {
            if(links[i].href.indexOf("melange.css") !== -1) {
                var otac = links[i].parentElement;
                var klinac = links[i];
                otac.removeChild(klinac);
                otac.appendChild(klinac);
            }
        }
    }

    window.melange.styles = styles;
})()
