//host = "http://45.55.210.25/";
host = "";

function initMenu() {
	$( ".cross" ).hide();
	$( ".menu" ).hide();
	$( ".hamburger" ).click(function() {
		$( ".menu" ).slideToggle( "slow", function() {
			$( ".hamburger" ).hide();
			$( ".cross" ).show();
		});
	});

	$( ".cross" ).click(function() {
		$( ".menu" ).slideToggle( "slow", function() {
			$( ".cross" ).hide();
			$( ".hamburger" ).show();
		});
	});
}				
		
function initMosaic(items, element) {	
	$(element).empty();

	for (i = 0; i < items.length; i++) { 
		addItemLayout(items[i], element);
	}
	
    $(element).masonry({
		itemSelector: '.grid-item',
		columnWidth: 160,
		isFitWidth: true
	});

	$(element).imagesLoaded().progress( function() {
		$(element).masonry('layout');
	});
}	

feed_done = false;
comments_done = false;
function initItems() {
	$.get(host + "services/items", function(data, success){
		if (feed_done) {
			$('#feed-grid').masonry('destroy');
		}
		initMosaic(data, '#feed-grid');		
		feed_done = true;	
		scaleLayout();
	});
	$.get(host + "services/bestComments", function(data, success){
		if (comments_done) {
			$('#comments-grid').masonry('destroy');
		}
		initMosaic(data, '#comments-grid');		
		comments_done = true;	
		scaleLayout();
	});
	
	$('a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
		var target = $(e.target).attr("href") 
		if (target == "#feed"){
			$('#feed-grid').masonry();
		} else if (target == "#best"){
			$('#comments-grid').masonry();
		}
				
	});
}	
		
	
function addItemLayout(item, element) {
	var extra;
	if (element == "#feed-grid"){
		extra = "-feed-like";
	} else if (element == "#comments-grid"){
		extra = "-best-like"
	}
	var cont;
	if (!item.b_comm) {
		cont = "<a href=\"item.html?" + getItemQuery(item.item_id, item.item_tid)
		+ "\"> <div class=\"grid-item\"><img class=\"image\" src=\"" + getExtUrl(item) + "\" title=\"" + getTitle(item).replace(/"/g, '&quot;')
		+ "\" onerror='this.onerror = null; this.src=\"" + getLocalUrl(item) + "\"'/></div> </a>";
	} else {
		var like_id = "#" + item.b_comm.id + extra;
		cont = "<div class=\"grid-item grid-item-large\">" +
				"<figure class=\"overlap-figure\">" +
					"<div class=\"item-container\">" +
						"<a href=\"item.html?" + getItemQuery(item.item_id, item.item_tid) + "\"> <img class=\"image image-large\" src=\"" + getExtUrl(item) + 
						"\" title=\"" + getTitle(item) + 
						"\" onerror='this.onerror = null; this.src=\"" + getLocalUrl(item) + "\"'/></div> </a>" +
						"<figcaption class=\"caption\">" +	
							"<div style=\"position:relative;\">" +	
								"<div class=\"like-box\">" 
									+ "<span id=\"" + item.b_comm.id + extra + "\">" + item.b_comm.likes + "</span>"
									+ "<a href=\"\" onclick=\"addLike('" + item.b_comm.id + "', '" + like_id + "'); return false;\"><img class=\"like-img\" src=\"media/like.png\"/></a>" +	
								"</div>" +
								"<div style=\"width:230px;\">" + item.b_comm.text + " (" + item.b_comm.author + ")" +
								"</div>" +								
							"</div>"
						"</figcaption>" +
					"</div>" +
				"</figure>" +
			"</div>";
	}
		
	$(element).append(cont);
}

function getExtUrl(item) {
	return item.img_url
}

function getLocalUrl(item) {
	if (item.item_id != null && item.item_id != "") {
		return host + "images/" + item.item_id + ".jpg";
	} else {
		return host + "images/temp/" + item.item_tid + ".jpg";
	}
}

function getTitle(item) {
	return item.title + " (" + item.source + ")";	
}

function getItemQuery(id, tid) {
	if (id != null && id != "" && id != 0) {
		return "item_id=" + id;
	} else {
		return "item_tid=" + tid;
	}
}


function showUserName() {
	var username = Cookies.get('name');
	if (username != null) {
		$(".username").html("Logged in as: <font color=\"red\">" + username + "</font>");
	} else {
		$(".username").html("");
	}
}

function addCommentLayout(comment) {	
	var like_id = "#" + comment.id + "-like";		
	var cont = "<div class=\"comment\">" +
				"<div class=\"like-box\">" +
					"<span id=\"" + comment.id + "-like\">" + comment.likes + "</span>" +
					"<a href=\"\" onclick=\"addLike('" + comment.id + "', '" + like_id + "'); return false;\"><img class=\"like-img\" src=\"media/like.png\"/></a>" +							
				"</div>" + 
				"<div class=\"comment-text\">" +
					"<span>" + comment.text + " (" + comment.author + ")</span>" +
				"</div>" +								
			"</div>";
	$('.comment-container').append(cont);
}

liking = false;
function addLike(id, span_id) {
	if (liking) return;
	liking = true;
	//var span_id = "#" + id + "-likes";
	var value = parseInt($(span_id).text());
	$(span_id).html(value+1);
	$.ajax({
		type: "GET",
		url: host + "services/like?comment_id=" + id,
		success: function(data){
			if (data != "OK"){
				$(span_id).html(value);
			}			
			liking = false;
		},
		error: function(XMLHttpRequest, textStatus, errorThrown) {			
			$(span_id).html(value);
			liking = false;
			if (XMLHttpRequest.status == 401) {
				logout();				
				window.location = host + "login.html";
			}
		}
	});
}

/*
function getItemInfo() {
	var item_tid = getQueryVariable("item_tid");
	var item_id = getQueryVariable("item_id");
	myitem = {item_tid: item_tid, item_id: item_id};	
	$.ajax({
		type: "GET",
		url: host + "services/itemInfo?" + getItemQuery(myitem),
		success: function(data){
			if (item_id == "" && data.item.item_id != null) {
				window.location.replace(host + "item.html?item_id=" + data.item.item_id);
				return;
			}
			$(".itemimg-caption").html(getTitle(data.item));
			$("#item-image").attr('src', getExtUrl(data.item));
			$("#item-image").attr('onerror', "this.onerror = null; this.src=\"" + getLocalUrl(myitem) + "\"");
			$('.comment-container').empty();
			if (data.comments != null) {
				for (i = 0; i < data.comments.length; i++) { 
					addCommentLayout(data.comments[i]);
				}
			}
			
		},
		error: function(XMLHttpRequest, textStatus, errorThrown) {
			
		}
	});
	
	var comment1 = {text: "Gran sketch", author: "Beppe", likes: 12, id: 100};
	var comment2 = {text: "Super sketch", author: "Leo", likes: 4, id: 101};
	var comments = [comment1, comment2];
	for (i = 0; i < comments.length; i++) { 
		addCommentLayout(comments[i]);
	}
	
}
*/

function postComment(id, tid) {
	$("#postcomment_result").html("");
	if ($("#comment").val() == "") {
		$("#postcomment_result").html("The comment can't be empty");
		return;
	}
	$("#spinner_item").show();
	$.ajax({
		type: "POST",
		url: host + "services/comment",
		data: "comment=" + $("#comment").val() + "&" + getItemQuery(id, tid),
		success: function(data){
			$("#spinner_item").hide();
			if (data == "OK"){
				//getItemInfo();	
				//window.location.reload(true);
				window.location.href = window.location.href;			
				$("#comment").val('');
				$("#postcomment_result").html("Thank you for posting!");	
			} else {
				$("#postcomment_result").html(data);	
			}		
		},
		error: function(XMLHttpRequest, textStatus, errorThrown) {
			$("#spinner_item").hide();			
			if (XMLHttpRequest.status == 401) {
				logout();
				initCommentForm();
			} else {
				$("#postcomment_result").html(textStatus);	
			}
		}
	});
}

function login() {
	$("#text_login").html("");
	$("#spinner_login").hide();
	if ($("#name_email").val() == "") {
		$("#text_login").html("Please insert a user name or an email account");
		return;
	}
	if ($("#key").val() == "") {
		$("#text_login").html("Please type your password");
		return;
	}
	$("#spinner_login").show();
			
	$.ajax({
		type: "POST",
		url: host + "services/login",
		data: $('form#login-form').serialize(),
		success: function(data){
				$("#spinner_login").hide();
				if (data == "OK") {
					window.location = host + "index.html";
				} else {
					$("#text_login").html(data);
				}
					
		},
		error: function(XMLHttpRequest, textStatus, errorThrown) {
			$("#spinner_login").hide();
			$("#text_login").html(textStatus);	
		}
	});
}
		
function recoverPass() {
	$( "#test_text" ).html("Recovering pass...");
}
		
function register() {
	$("#text_register").html("");
	$("#spinner_register").hide();
	if ($("#name").val() == "") {
		$("#text_register").html("Please choose a user name");
		return;
	}
	if ($("#email").val() == "") {
		$("#text_register").html("Please choose an email account");
		return;
	}
	if ($("#pass_first").val() == "") {
		$("#text_register").html("Please choose a password");
		return;
	}
	if ($("#pass_first").val() != $("#pass_second").val()) {
		$("#text_register").html("Passwords do not match");
		return;
	}
	$("#spinner_register").show();
			
	$.ajax({
		type: "POST",
		url: host + "services/register",
		data: $('form#register-form').serialize(),
		success: function(data){
			$("#spinner_register").hide();
			if (data == "OK"){
				$("#text_register").html("Registration successful. An email has been sent to your account for verification");
			} else if (data == "ERROR_NAME_USED") {
				$("#text_register").html("Error: The name is already used");
			} else if (data == "ERROR_EMAIL_USED") {
				$("#text_register").html("Error: The email is already used");
			} else if (data == "ERROR_DB") {
				$("#text_register").html("Error");
			} else if (data == "ERROR_EMAIL_INVALID") {
				$("#text_register").html("Error: invalid email");
			} else if (data == "ERROR_FORMAT") {
				$("#text_register").html("Error: wrong format");
			} else {
				$("#text_register").html(data);
			}			
		},
		error: function(XMLHttpRequest, textStatus, errorThrown) {
			$("#spinner_register").hide();
			$("#text_register").html(textStatus);	
		}
	});			
}

function logout() {
	Cookies.remove('name', { path: '/' });
	Cookies.remove('token', { path: '/' });
	$( ".cross" ).hide();
	$( ".hamburger" ).show();
	$( ".menu" ).hide();
	showUserName();
}

/*	
function getParameterByName(name) {
    name = name.replace(/[\[]/, "\\[").replace(/[\]]/, "\\]");
    var regex = new RegExp("[\\?&]" + name + "=([^&#]*)"),
    results = regex.exec(location.search);
    return results === null ? "" : decodeURIComponent(results[1].replace(/\+/g, " "));
}
*/

function getQueryVariable(variable) {
	var query = window.location.search.substring(1);
	var vars = query.split("&");
	for (var i=0;i<vars.length;i++) {
		var pair = vars[i].split("=");
		if(pair[0] == variable){return pair[1];}
	}
	return(false);
}

function setResponsive() {
	scaleLayout();
	$(window).bind("load", scaleLayout);
	$(window).bind("resize", scaleLayout);
	$(window).bind("orientationchange", scaleLayout);
}


function initCommentForm() {
	$("#spinner_item").hide();
	var username = Cookies.get('name');			
	if (username != null){
		$(".commentform").show();
		$("#signin_link").hide();
	} else {
		$(".commentform").hide();
		$("#signin_link").show();
	}
}
