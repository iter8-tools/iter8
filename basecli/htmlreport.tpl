<!doctype html>
<html lang="en">

  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">

    <!-- Font Awesome -->
    <script src="https://kit.fontawesome.com/db794f5235.js" crossorigin="anonymous"></script>

    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">

    <style>
      html {
        font-size: 18px;
      }		
    </style>

    <title>Iter8 Experiment Report</title>
  </head>

  <body>

    <!-- jQuery first, then Popper.js, then Bootstrap JS -->
    <script src="https://code.jquery.com/jquery-3.3.1.slim.min.js" integrity="sha384-q8i/X+965DzO0rT7abK41JStQIAqVgRVzpbzo5smXKp4YfRvH+8abtTE1Pi6jizo" crossorigin="anonymous"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.7/umd/popper.min.js" integrity="sha384-UO2eT0CpHqdSJQ6hJty5KVphtPhzWj9WO1clHTMGa3JDZwrnQq4sF86dIHNDz0W1" crossorigin="anonymous"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>
  
    <!-- NVD3 -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/nvd3@1.8.6/build/nv.d3.css">
    <!-- Include d3.js first -->
    <script src="https://cdn.jsdelivr.net/npm/d3@3.5.3/d3.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/nvd3@1.8.6/build/nv.d3.js"></script>

    <div class="container">

      <h1 class="display-4">Experiment Report</h1>
      <h3 class="display-6">Insights from Iter8 Experiment</h3>
      <hr>

      <div class="toast fade {{ htmlRenderStrVal . "showClass" }} mw-100" role="alert" aria-live="assertive" aria-atomic="true">
        <div class="toast-header {{ htmlRenderStrVal . "textColor" }}">
          <strong class="mr-auto">
            Experiment Status
            &nbsp;&nbsp;
            <i class="fas fa-thumbs-{{ htmlRenderStrVal . "thumbs" }}"></i>
          </strong>
          <button type="button" class="ml-2 mb-1 close" data-dismiss="toast" aria-label="Close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="toast-body {{ htmlRenderStrVal . "textColor" }}">
          {{ htmlRenderStrVal . "msg" }}
        </div>
      </div>
    
      <script>
      $(document).ready(function(){
        $(function () {
          $('[data-toggle="tooltip"]').tooltip()
        });	

        $(".toast").toast({
          autohide: false,
          delay: 10000
        }).toast('show');
      });	
      </script>

    </div>
  </body>
</html>
