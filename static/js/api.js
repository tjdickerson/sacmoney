
function add_transaction() {
	const trans_date = document.getElementById("input-trans-date").value;
	const trans_name = document.getElementById("input-trans-name").value;
	const trans_amount = document.getElementById("input-trans-amount").value;
	const form = document.getElementById("form-new-transaction")
	console.log(form)

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
			day: recurring_date,
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
	const error_display = document.getElementById("error-display");
	const error_text = document.getElementById("error-text");

	error_display.style.display = "block";
	error_text.innerHTML = error;

	setTimeout(() => {
		error_display.style.display = "none";
	}, 5000)
}


function delete_transaction(sender) {
	const trn_id = sender.getAttribute("tid");
	post("/deleteTransaction",
		(rt) => { after_post(rt) },
		{
			Id: trn_id,
		});
}

function delete_recurring_transaction(sender) {
	const recurr_id = sender.getAttribute("rid");
	post("/deleteRecurring",
		(rt) => { after_post(rt) },
		{
			Id: recurr_id,
		});
}

/**
 * @param {string} uri
 * @param {function} callback
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

/**
 * @param {string} uri
 * @param {function} callback
 * @param {JSONObject} data
 * */
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

function page_load_transactions() {
	const input_date = document.getElementById("input-trans-date");
	input_date.valueAsDate = new Date();

	const input_name = document.getElementById("input-trans-name");
	const input_amount = document.getElementById("input-trans-amount");

	input_name.value = "";
	input_amount.value = "";

	input_name.focus();

	set_default_button(input_amount);
}

function page_load_recurrings() {
	const input_date = document.getElementById("input-recurring-date");
	const input_name = document.getElementById("input-recurring-name");
	const input_amount = document.getElementById("input-recurring-amount");

	input_date.value = "";
	input_name.value = "";
	input_amount.value = "";

	input_date.focus();

	set_default_button(input_amount);
}

function page_load_accounts() {
	const input_name = document.getElementById("input-account-name");

	input_name.value = "";
	input_name.focus();

	set_default_button(input_name);
}

/**
* @param {HTMLElement} target_el
**/
function set_default_button(target_el) {
	const add_btn = document.getElementById("btn-add");
	target_el.addEventListener("keypress", function(event) {
		if (event.key === "Enter") {
			event.preventDefault();
			add_btn.click();
		}
	});
}

