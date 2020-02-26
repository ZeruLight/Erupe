function createErrorAlert(message) {
	parent.postMessage(message, "*");
}


// Function to continually check if we got a login result yet,
// then navigating to the character selection if we did.
function checkAuthResult() {
    var loginResult = window.external.getLastAuthResult();
    console.log('|' + loginResult + '|');
    if(loginResult == "AUTH_PROGRESS") {
        setTimeout(checkAuthResult, 500);
    } else if (loginResult == "AUTH_SUCCESS") {
        window.location.href = 'charsel.html'
    } else {
        createErrorAlert("Error logging in!");
    }
}

$(function() {
    // Login form submission.
    $("#loginform").submit(function(e){
        e.preventDefault();

        username = $("#username").val();
        password = $("#password").val();

        try{
            window.external.loginCog(username, password, password);
        } catch(e){
            createErrorAlert("Error on loginCog: " + e);
        }

        checkAuthResult();
    });

    // Config button.
    $("#configButton").click(function(){
        try{
            window.external.openMhlConfig();
        } catch(e){
            createErrorAlert("Error on openMhlConfig: " + e);
        }
    });
});
