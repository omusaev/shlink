
$( document ).ready(function() {
    $( "#shlink_form" ).submit(function( event ) {
        event.preventDefault();

        var backend_url = "/";
        var source_link = $('#source_link').val();

        $.ajax({
            method: "POST",
            url: backend_url,
            data: { Source: source_link },
            success: function(data, textStatus, jqXHR) {
                var shlink = document.URL + 'r/' + data.shlink_name;
                $('#shlink').val(shlink);
            }
        })
    });
});