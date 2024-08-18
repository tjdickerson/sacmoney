
function page_load() {
	const input_date = document.getElementById("input-trans-date");
	input_date.valueAsDate = new Date();
	reset_fields();
	load_transactions();
	handle_listeners();
}

function load_transactions() {
	request("/getTransactions", (responseText) => {
		let trans_container = document.getElementById("transactions");
		trans_container.innerHTML = responseText;
	});
}

function add_transaction() {
	const trans_date = document.getElementById("input-trans-date").value;
	const trans_name = document.getElementById("input-trans-name").value;
	const trans_amount = document.getElementById("input-trans-amount").value;

	post("/addTransaction",
		(rt) => { post_add_transaction(rt); },
		{
			date: trans_date,
			name: trans_name,
			amount: trans_amount,
		});
}

function post_add_transaction(result) {
	if (!result.indexOf("::") < 0) {
		console.log(`Error: ${result}`);
		return;
	}

	const res = result.split("::");
	const status = res[0];
	const newAvailable = res[1];
	const divAvail = document.getElementById("account-available");

	console.log(status);
	divAvail.innerText = newAvailable;

	load_transactions();
	reset_fields();
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

function post(uri, callback, data) {
	const xhr = new XMLHttpRequest();
	xhr.open("POST", uri);
	xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
	xhr.onload = () => {
		if (xhr.readyState == 4 && xhr.status == 200) {
			callback(xhr.responseText);
		} else {
			console.error(`Post failed: ${xhr.status}`);
		}
	}

	let send_data = JSON.stringify(data);
	console.log(`sending: ${send_data}`);
	xhr.send(send_data);
}


function handle_listeners() {
	const input_amount = document.getElementById("input-trans-amount");
	const add_btn = document.getElementById("btn-add");

	input_amount.addEventListener("keypress", function(event) {
		if (event.key === "Enter") {
			event.preventDefault();
			add_btn.click();
		}
	});
}


function reset_fields() {
	const input_name = document.getElementById("input-trans-name");
	const input_amount = document.getElementById("input-trans-amount");

	input_name.value = "";
	input_amount.value = "";

	input_name.focus();
}



