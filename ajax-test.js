
  function ajax(url, method, success, failure) {
    var xmlHttpReq = false;
    var self = this;
    // Mozilla/Safari
    if (window.XMLHttpRequest) {
        self.xmlHttpReq = new XMLHttpRequest();
    }
    // IE
    else if (window.ActiveXObject) {
        self.xmlHttpReq = new ActiveXObject("Microsoft.XMLHTTP");
    }
    self.xmlHttpReq.open(method, strURL, true);
    self.xmlHttpReq.responseType = "json";
    self.xmlHttpReq.onreadystatechange = function() {
        if (self.xmlHttpReq.readyState == 4) {
            if (self.xmlHttpReq.status != 200) {
              failure(self.xmlHttpReq.status, self.xmlHttpReq.response);
            } else {
              success(self.xmlHttpReq.responseText);
            }
        }
    }
    self.xmlHttpReq.send();
  }
