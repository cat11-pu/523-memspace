(function() {

  $(function() {
    $(window).on('hashchange', function() {
      switch (location.hash) {
      case '':
      case '#assembler':
        document.body.className = 'showing-assembler';
        break;
      case '#disassembler':
        document.body.className = 'showing-disassembler';
        break;
      case '#debugger':
        document.body.className = 'showing-debugger';
        break;
      }
    }).trigger('hashchange');
  });

})();
