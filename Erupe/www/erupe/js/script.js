var __mhf_launcher = {};
var loginScreen = true;
var doingAuto = false;
var uids;
var selectedUid;
var firstChar;
var modalState = false;


function soundSel() {
	window.external.playSound('IDR_WAV_SEL');
}

function soundOk() {
	window.external.playSound('IDR_WAV_OK');
}

function soundPreLogin() {
	window.external.playSound('IDR_WAV_PRE_LOGIN');
}

function soundLogin() {
	window.external.playSound('IDR_WAV_LOGIN');
}

function soundNiku() {
	window.external.playSound('IDR_NIKU');
}

function addLog(text, mode) {
  switch (mode) {
    case 'winsock':
      text = '<span class="winsock">'+text+'</span><br>';
      break;
    case 'normal':
      text = '<span class="white">'+text+'</span><br>';
      break;
    case 'good':
      text = '<span class="green">'+text+'</span><br>';
      break;
    case 'error':
      text = '<span class="red">'+text+'</span><br>';
      break;
  }
  let logText = document.getElementById('log_p');
  logText.innerHTML = logText.innerHTML + text;
  let logBox = document.getElementsByClassName('log_inner')[0];
  logBox.scrollTop = logBox.scrollHeight;
}

function loadAccount() {
  let allowed = localStorage.getItem('saving');
  if (allowed != 'null' && allowed == 'true') {
    document.getElementById('username').value = localStorage.getItem('username');
    document.getElementById('password').value = localStorage.getItem('password');
    document.getElementById('login_save').checked = true;
    let autoEnabled = localStorage.getItem('autologin');
    if (autoEnabled != 'null' && autoEnabled == 'true') {
      doingAuto = true;
      doLogin();
    }
  }
}

function saveAccount() {
	let checkbox = document.getElementById('login_save');
	if (checkbox.checked == true) {
    let username = document.getElementById('username').value;
  	let password = document.getElementById('password').value;
		if (username[username.length - 1] == '+') {
			username = username.slice(0, username.length - 1)
		}
		localStorage.setItem('username', username);
		localStorage.setItem('password', password);
		localStorage.setItem('saving', 'true');
	} else {
		localStorage.removeItem('username');
		localStorage.removeItem('password');
		localStorage.removeItem('saving');
		localStorage.removeItem('uid');
	}
}

function createCharItem(name, uid, weapon, hr, gr, date, sex) {
  var icon;
  const dateObject = new Date(date * 1000);
  date = dateObject.toLocaleDateString('en-US');
  let dateString = '';
  for (var i = 0; i < date.length; i++) {
    if (date[i] != '‎') { // invisible LTR char
      dateString += date[i];
    }
  }
  if (sex == 'M') {
    sex = "♂";
  } else {
    sex = "♀";
  }
  if (hr > 999) {
	hr = 999;
  }
  if (gr > 999) {
	gr = 999;
  }
  switch (weapon) {
    case '片手剣':
      weapon = 'Sword & Shield';
      icon = 'img/icons/ss.png';
      break;
    case '双剣':
      weapon = 'Dual Blades';
      icon = 'img/icons/db.png';
      break;
    case '大剣':
      weapon = 'Greatsword';
      icon = 'img/icons/gs.png';
      break;
    case '太刀':
      weapon = 'Long Sword';
      icon = 'img/icons/ls.png';
      break;
    case 'ハンマー':
      weapon = 'Hammer';
      icon = 'img/icons/hm.png';
      break;
    case '狩猟笛':
      weapon = 'Hunting Horn';
      icon = 'img/icons/hh.png';
      break;
    case 'ランス':
      weapon = 'Lance';
      icon = 'img/icons/ln.png';
      break;
    case 'ガンランス':
      weapon = 'Gunlance';
      icon = 'img/icons/gl.png';
      break;
    case '穿龍棍':
      weapon = 'Tonfa';
      icon = 'img/icons/tf.png';
      break;
    case 'スラッシュアックスF':
      weapon = 'Switch Axe F';
      icon = 'img/icons/sa.png';
      break;
    case 'マグネットスパイク':
      weapon = 'Magnet Spike';
      icon = 'img/icons/ms.png';
      break;
    case 'ヘビィボウガン':
      weapon = 'Heavy Bowgun';
      icon = 'img/icons/hbg.png';
      break;
    case 'ライトボウガン':
      weapon = 'Light Bowgun';
      icon = 'img/icons/lbg.png';
      break;
    case '弓':
      weapon = 'Bow';
      icon = 'img/icons/bow.png';
      break;
    default:
      weapon = 'Unknown';
      icon = 'img/icons/uk.png';
  }
  let charElem = document.createElement('DIV');
  charElem.setAttribute('href', '#');
  charElem.id = uid;
  charElem.classList.add('unit');
  if (firstChar) {
    firstChar = false;
    selectedUid = uid;
    charElem.classList.add('active');
  }

  let elemName = document.createElement('DIV');
  elemName.id = 'char_name';
  elemName.innerHTML = name;
  charElem.appendChild(elemName);
  let elemWeapon = document.createElement('DIV');
  elemWeapon.id = 'char_weapon';
  elemWeapon.innerHTML = weapon;
  charElem.appendChild(elemWeapon);
  let elemHr = document.createElement('DIV');
  elemHr.id = 'char_hr';
  elemHr.innerHTML = 'HR'+hr;
  charElem.appendChild(elemHr);
  let elemGr = document.createElement('DIV');
  elemGr.id = 'char_gr';
  elemGr.innerHTML = 'GR'+gr;
  charElem.appendChild(elemGr);
  let elemSex = document.createElement('DIV');
  elemSex.id = 'char_sex';
  elemSex.innerHTML = sex;
  charElem.appendChild(elemSex);
  let elemUid = document.createElement('DIV');
  elemUid.id = 'char_uid';
  elemUid.innerHTML = 'ID: '+uid;
  charElem.appendChild(elemUid);
  let elemLastLogin = document.createElement('DIV');
  elemLastLogin.id = 'char_login';
  elemLastLogin.innerHTML = 'Last Login: '+dateString;
  charElem.appendChild(elemLastLogin);

  let iconElem = document.createElement('IMG');
  iconElem.src = icon;
  charElem.appendChild(iconElem);
  let unitsElem = document.getElementById('units');
  unitsElem.appendChild(charElem);
}

function switchPrompt() {
  loginScreen = !loginScreen;
  if (loginScreen) {
    document.getElementById('units').innerHTML = '';
    document.getElementById('char_select').style.display = 'none';
    document.getElementById('login').style.display = 'block';
  } else { // Character selector
    document.getElementById('login').style.display = 'none';
    document.getElementById('char_select').style.display = 'block';
    try {
      // Example data for browser testing
      //var charInfo = "<?xml version='1.0' encoding='shift_jis'?><CharacterInfo defaultUid=''><Character name='Cynthia' uid='211111' weapon='双剣' HR='7' GR='998' lastLogin='1645961490' sex='F' /><Character name='狩人申請可能' uid='311111' weapon='大剣' HR='7' GR='0' lastLogin='1650486502' sex='M' /></CharacterInfo>";
      var charInfo = window.external.getCharacterInfo();
      charInfo = charInfo.split("'").join('"');
      charInfo = charInfo.split('&apos;').join("'");
    } catch (e) {
      addLog('Error getting character info: '+e, 'error');
    }
    try {
      firstChar = true;
      uids = new Array();
      parser = new DOMParser();
      let xml = parser.parseFromString(charInfo, 'text/xml');
      let numChars = xml.getElementsByTagName('Character').length;
      for (var i = 0; i < numChars; i++) {
        let char = xml.getElementsByTagName('Character')[i].attributes;
        createCharItem(
          char.name.value,
          char.uid.value,
          char.weapon.value,
          char.HR.value,
          char.GR.value,
          char.lastLogin.value,
          char.sex.value
        );
        uids.push(char.uid.value);
      }
  	} catch (e) {
      addLog('Error parsing character info XML: '+e, 'error');
  	}
    let uid = localStorage.getItem('uid');
    if (uid != 'null' && uids.indexOf(uid) >= 0) {
      setUidIndex(uids.indexOf(uid));
    }
  }
}

function doLogin(option) {
  let username = document.getElementById('username').value;
  let password = document.getElementById('password').value;
  if (username == '') {
    addLog('Please enter Erupe ID!', 'error');
  } else if (password == '') {
    addLog('Please enter Password!', 'error');
  } else {
    document.getElementById('processing').style.display = 'block';
    soundPreLogin();
    addLog('Authenticating...', 'normal');
    try {
      if (option) {
        addLog('Creating new character...', 'normal');
        window.external.loginCog(username+'+', password, password);
      } else {
        window.external.loginCog(username, password, password);
	  }
    } catch (e) {
      addLog('Error on loginCog: '+e, 'error');
    }
    checkAuth();
  }
}

function checkAuth() {
  let loginResult = window.external.getLastAuthResult();
  if (loginResult == 'AUTH_PROGRESS') {
    setTimeout(checkAuth, 10);
    return;
  } else if (loginResult == 'AUTH_SUCCESS') {
    saveAccount();
    addLog('Connected.', 'good');
    if (doingAuto) {
			let uid = localStorage.getItem('uid');
			window.external.selectCharacter(uid, uid);
			window.external.exitLauncher();
		} else {
      addLog('After selecting a character, press [Launch]', 'normal');
      switchPrompt();
		}
  } else {
    addLog('Error logging in: '+loginResult+':'+window.external.getSignResult(), 'error');
  }
  document.getElementById('processing').style.display = 'none';
}

function launch() {
  document.getElementById('game_starting').style.display = 'block';
  try {
    window.external.selectCharacter(selectedUid, selectedUid);
  } catch (e) {
    addLog('Error selecting character: '+e, 'error');
    document.getElementById('game_starting').style.display = 'none';
  }
  let allowed = localStorage.getItem('saving');
  if (allowed != 'null' && allowed == 'true') {
    localStorage.setItem('uid', selectedUid);
    let autoBox = document.getElementById('auto_box');
    if (autoBox.checked) {
      localStorage.setItem('autologin', true);
    }
  }
  setTimeout(function () {
    window.external.exitLauncher();
  }, 3000);
}

function autoWarning() {
  let autoBox = document.getElementById('auto_box');
  if (autoBox.checked) {
    addLog('Auto-Login is for advanced users, to disable it you will need to clear your IE cache. Uncheck the box now if you are not an advanced user.', 'error');
  }
}

function charselScrollUp() {
  let index = uids.indexOf(selectedUid) - 1;
  if (index < 0) {
    index = uids.length - 1;
  }
  setUidIndex(index);
}

function charselScrollDown() {
  let index = uids.indexOf(selectedUid) + 1;
  if (index == uids.length) {
    index = 0;
  }
  setUidIndex(index);
}

function setUidIndex(index) {
  let units = document.getElementsByClassName('unit');
  let numUnits = units.length;
  for (var i = 0; i < numUnits; i++) {
    units[i].classList.remove('active');
  }
  selectedUid = uids[index];
  document.getElementById(selectedUid).classList.add('active');
}

function toggleModal(preset, url) {
  let modal = document.getElementById('launcher_modal');
  modalState = !modalState;
  if (modalState) {
    setModalContent(preset, url);
    modal.style.display = 'block';
  } else {
    modal.style.display = 'none';
  }
}

function setModalContent(preset, url) {
  let modal = document.getElementById('launcher_modal');
	switch (preset) {
		case 'openLink':
			modal.querySelector('.dialog p').innerHTML = ' \
				Are you sure you want to open this URL? \
				<br> \
				<span class="uid">'+url+'</span> \
				<br> \
				<div class="sp"></div> \
				<span class="attention">This will open in a browser</span> \
			';
			modal.querySelector('.dialog .btns').innerHTML = ' \
				<ul> \
					<li> \
						<div onmouseover="soundSel()" onclick="soundOk(); window.external.openBrowser(\''+url+'\'); toggleModal(0)">Open</div> \
					</li> \
					<li> \
						<div onmouseover="soundSel()" onclick="soundOk(); toggleModal(0)">Cancel</div> \
					</li> \
				</ul> \
			';
			break;
		case 'confirmCharDelete':
			modal.querySelector('.dialog p').innerHTML = ' \
				Are you sure you want to delete your character? \
				<br>NAME \
				<span class="uid"> (ID: 000000)</span> \
				<br> \
				<div class="sp"></div> \
				<span class="attention">You will not be able to recover this character, \
					<br>it will be gone forever. \
				</span> \
			';
			modal.querySelector('.dialog .btns').innerHTML = ' \
				<ul> \
					<li> \
						<div onmouseover="soundSel()" onclick="soundOk(); addLog(\'Not yet implemented.\', \'error\'); toggleModal(0)">Yes</div> \
					</li> \
					<li> \
						<div onmouseover="soundSel()" onclick="soundOk(); toggleModal(0)">Cancel</div> \
					</li> \
				</ul> \
			';
			// Uses the launcher delete
			// modal.querySelector(".dialog .btns").innerHTML = "<ul><li><div unselectable=\"on\" onselectstart=\"return false;\" onmouseover=\"soundSel();\" onclick=\"soundOk(); window.external.deleteCharacter('"+selectedUid+"');  toggleModal(0);\" style=\"opacity: 1;\">Yes</div></li><li><div onmouseover=\"soundSel();\" onclick=\"soundOk(); toggleModal(0);\" unselectable=\"on\" onselectstart=\"return false;\" class=\"\">Cancel</div></li></ul>";
			break;
		case 'addCharNew':
			modal.querySelector('.dialog p').innerHTML = ' \
				Are you sure you want to add a new character? \
				<br> \
				<div class="sp"></div> \
				<span class="attention">Press [Add Character] to add a new slot.</span> \
			';
			modal.querySelector('.dialog .btns').innerHTML = ' \
				<ul> \
					<li> \
						<div onmouseover="soundSel()" onclick="soundNiku(); doLogin(1); switchPrompt(); toggleModal(0)">Add Character</div> \
					</li> \
					<li> \
						<div onmouseover="soundSel()" onclick="soundOk(); toggleModal(0)">Cancel</div> \
					</li> \
				</ul> \
			';
			break;
		default:
			return;
	}
}

function charselAdd() {
  toggleModal('addCharNew');
}

function charselDel() {
  toggleModal('confirmCharDelete');
}

function charselLog() {
  addLog('Disconnected.', 'error');
  addLog('Enter Erupe ID and Password, then press [Log In]', 'normal');
  switchPrompt();
}

function doEval() {
  try {
    addLog(eval(document.getElementById('console').value), 'error');
  } catch (e) {
    addLog('Error on doEval: '+e, 'error');
  }
}

function init() {
  document.addEventListener('keypress', function(e) {
    if (e.key == '~') {
      document.getElementById('dev').style.display = 'block';
    }
  });
  let unselectable = document.getElementsByClassName('unselectable');
  let unselectableLen = unselectable.length;
  for (var i = 0; i < unselectableLen; i++) {
    unselectable[i].setAttribute('unselectable', 'on');
    unselectable[i].addEventListener('selectstart', function(){return false;});
    unselectable[i].addEventListener('mouseover', function(){window.external.beginDrag(false);});
  }
  let grabbable = document.getElementsByClassName('grabbable');
  let grabbableLen = grabbable.length;
  for (var i = 0; i < grabbableLen; i++) {
    grabbable[i].addEventListener('selectstart', function(){window.external.beginDrag(true);});
    grabbable[i].addEventListener('mousedown', function(){window.external.beginDrag(true);});
    grabbable[i].addEventListener('mouseup', function(){window.external.beginDrag(true);});
  }
  document.getElementById('login_save_text').addEventListener('click', function() {
    let checkbox = document.getElementById('login_save');
    checkbox.checked = !checkbox.checked;
  });
  document.getElementById('auto_text').addEventListener('click', function() {
    let checkbox = document.getElementById('auto_box');
    checkbox.checked = !checkbox.checked;
  });
  document.getElementById('username').focus();
  loadAccount();
  addLog('Winsock Ver. [2.2]', 'winsock');
  addLog('Enter Erupe ID and Password, then press [Log In]', 'normal');
}

init();