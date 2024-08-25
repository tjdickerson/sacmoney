
function page_load_transactions() {
	const input_date = document.getElementById("input-trans-date");
	input_date.valueAsDate = new Date();
	reset_fields();
	handle_listeners();
}

function page_load_recurrings() {
}

function page_load_accounts() {
}

function add_transaction() {
	const trans_date = document.getElementById("input-trans-date").value;
	const trans_name = document.getElementById("input-trans-name").value;
	const trans_amount = document.getElementById("input-trans-amount").value;

	post("/addTransaction",
		(rt) => { after_post(rt) },
		{
			id: "0",
			date: trans_date,
			name: trans_name,
			amount: trans_amount,
		});
}

function add_recurring_transaction() {
	const recurring_date = document.getElementById("input-recurring-date").value;
	const recurring_name = document.getElementById("input-recurring-name").value;
	const recurring_amount = document.getElementById("input-recurring-amount").value;

	post("/addRecurring",
		(rt) => { after_post(rt); },
		{
			date: recurring_date,
			name: recurring_name,
			amount: recurring_amount,
		});
}

function add_account() {
	const account_name = document.getElementById("input-account-name").value;

	post("/addAccount",
		(rt) => { after_post(rt); },
		{
			name: account_name,
		});
}

function after_post(result) {
	console.log(result)
	if (result === "SUCCESS") {
		window.location.reload();
	} else {
		show_error(result);
	}
}

function show_error(error) {
	alert(error);
}


function delete_transaction(sender) {
	const trn_id = sender.getAttribute("tid");
	post("/deleteTransaction",
		(rt) => { post_change_transaction(rt) },
		{
			Id: trn_id,
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



