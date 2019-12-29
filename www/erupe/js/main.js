// Helper function to dynamically create a bootstrap alert box.
function createErrorAlert(message) {
    var tmpDiv = $('<div/>')
    	.attr("id", "myAlertBoxID")
      .attr("role", "alert")
      .addClass("alert alert-danger alert-dismissible fade show")
    
    tmpDiv.append(message);
    tmpDiv.append($("<button/>")
    	.attr("type", "button")
      .addClass("close")
      .attr("data-dismiss", "alert")
      .attr("aria-label", "Close")
      .append($("<span/>")
      	.attr("aria-hidden", "true")
      	.text("×")
      ));
    
   $("#alertBox").append(tmpDiv);
}

function doLauncherInitalize() {
    try{
        window.external.getMhfMutexNumber();
    } catch(e){
        createErrorAlert("Error getting Mhf mutex number! " + e);
    }

    try{
        var serverListXml = window.external.getServerListXml();
    } catch(e){
        createErrorAlert("Error getting serverlist.xml! " + e);
    }

    if(serverListXml == ""){
        createErrorAlert("Got empty serverlist.xml!");
    }
    console.log(serverListXml);

    try{
        var lastServerIndex = window.external.getIniLastServerIndex();
    } catch(e){
        createErrorAlert("Error on getIniLastServerIndex: " + e);
    }
    console.log("Last server index:" + lastServerIndex);

    try{
        window.external.setIniLastServerIndex(0);
    } catch(e){
        createErrorAlert("Error on setIniLastServerIndex: " + e);
    }

    try{
        var mhfBootMode = window.external.getMhfBootMode();
    } catch(e){
        createErrorAlert("Error on getMhfBootMode: " + e);
    }
    console.log("mhfBootMode:" + mhfBootMode);

    try{
        var userId = window.external.getUserId();
    } catch(e){
        createErrorAlert("Error on getUserId: " + e);
    }
    console.log("userId:" + userId);

    try{
        var password = window.external.getPassword();
    } catch(e){
        createErrorAlert("Error on getPassword: " + e);
    }
    console.log("password:" + password);
}

$(function() {
    // Setup the titlebar and exit button so that the window works how you would expect.
    $("#titlebar").on("click", function(e) {
        window.external.beginDrag(true);
    });

    $("#exit").on("click", function(e) {
        window.external.closeWindow();
    });

    // Setup the error message passthrough
    $(window).on("message onmessage", function(e) {
        var data = e.originalEvent.data;
        createErrorAlert(data)
    });

    // Initialize the launcher by calling the native/external functions it exposes in the proper order.
    doLauncherInitalize();
});