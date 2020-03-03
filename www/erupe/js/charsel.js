function createErrorAlert(message) {
	parent.postMessage(message, "*");
}

function createCharListItem(name, uid, weapon, HR, GR, lastLogin, sex) {
    var topDiv = $('<div/>')
        .attr("href", "#")
        .attr("uid", uid)
        .addClass("char-list-entry list-group-item list-group-item-action flex-column align-items-start");

    var topLine = $('<div/>')
            .addClass("d-flex w-100 justify-content-between")
            .append(
                $('<h5/>')
                .addClass("mb-1")
                .text(name)
            )
            .append(
                $('<small/>')
                .text('ID:' + uid)
            );

    var bottomLine = $('<div/>')
        .addClass("d-flex w-100 justify-content-between")
        .append($('<small/>').text('Weapon: ' + weapon))
        .append($('<small/>').text('HR: ' + HR))
        .append($('<small/>').text('GR: ' + GR))
        .append($('<small/>').text('LastLogin: ' + lastLogin))
        .append($('<small/>').text('Sex: ' + sex));

    topDiv.append(topLine);
    topDiv.append(bottomLine);
    
   $("#characterlist").append(topDiv);
}

$(function() {
    try {
        var charInfo = window.external.getCharacterInfo();
    } catch (e) {
        createErrorAlert("Error on getCharacterInfo!");
    }
    
    try{
        // JQuery's parseXML isn't working properly on IE11, use the activeX XMLDOM instead.
        //$xmlDoc = $.parseXML(charInfo),
        $xmlDoc = new ActiveXObject("Microsoft.XMLDOM");
        $xmlDoc.async = "false";
        $xmlDoc.loadXML(charInfo);
        
        $xml = $($xmlDoc);
    } catch (e) {
        createErrorAlert("Error parsing character info xml!" + e);
    }

    // Go over each "Character" element in the XML and then create a new list item for it.
    try {
        $($xml).find("Character").each(function(){
            createCharListItem(
                $(this).attr('name'),
                $(this).attr('uid'),
                $(this).attr('weapon'),
                $(this).attr('HR'),
                $(this).attr('GR'),
                $(this).attr('lastLogin'),
                $(this).attr('sex')
            );
        });
    } catch (e) {
        createErrorAlert("Error searching character info xml!");
    }

    // Set the active character selection on click.
    $(".char-list-entry").click(function(){
        if(!$(this).hasClass("active")) {
            $(".char-list-entry.active").removeClass("active");
            $(this).addClass("active");
        }
    });

    $("#selectButton").on("click", function() {
        // Get the active character selection.
        var selectedUid = $(".char-list-entry.active").attr("uid");

        // Call into the native launcher select function.
        try{
            window.external.selectCharacter(selectedUid, selectedUid)
        } catch(e) {
            createErrorAlert("Error on select character!");
        }
        
        // If we didn't error before, just close the launcher to start the game.
        setTimeout(function(){
            window.external.exitLauncher();
        }, 500);
    });
		
	$("#configButton").click(function() {
        try{
            window.external.openMhlConfig();
        } catch(e){
            createErrorAlert("Error on openMhlConfig: " + e);
        }
    })

});
