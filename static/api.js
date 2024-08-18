
function clicky() {
	loadTransactions();
}

function loadTransactions() {
	request("/getTransactions", (responseText) => {
		let trans_container = document.getElementById("transaction-container");
		trans_container.innerHTML = responseText;
	});
}

/**
 * @param {string} uri
 * @param {function} callback
 * @returns {string}
 * */
function request(uri, callback) {
	const xhr = new XMLHttpRequest();
	xhr.open("POST", uri);
	xhr.setRequestHeader("Content-Type", "application/text");
	xhr.onload = () => {
		if (xhr.readyState == 4 && xhr.status == 200) {
			callback(xhr.responseText);
		} else {
			console.error(`Request failed: ${xhr.status}`);
		}
	}
	xhr.send();
}
