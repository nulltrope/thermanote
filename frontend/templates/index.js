const toolbarOptions = {
    container: '#toolbar',
    // handlers: {
    //   // handlers object will be merged with default handlers object
    //   link: function (value) {
    //     const text = this.quill.getSemanticHTML();
    //     console.log(text);
    //     console.log("Sending request to backend!");
    //     try {
    //         fetch("http://localhost:8001/submit", {
    //             method: "POST",
    //             body: text,
    //         }).then((resp) => {
    //             if (!resp.ok) {
    //                 console.log("error:", resp.statusText);
    //             } else {
    //                 resp.text().then((rawText) => console.log(rawText));
    //             }
    //         });
    //     } catch (err) {
    //         console.log("ERROR: ", err);
    //     }
    //   },
    //   clean: function (value) {
    //     const text = this.quill.getSemanticHTML();
    //     console.log(text);
    //     console.log("Sending request to backend!");
    //     try {
    //         fetch("http://localhost:8001/preview", {
    //             method: "POST",
    //             body: text,
    //         }).then((resp) => {
    //             if (!resp.ok) {
    //                 console.log("error:", resp.statusText);
    //             } else {
    //                 resp.blob().then((blob) => {
    //                   const reader = new FileReader();
    //                   reader.readAsDataURL(blob);
    //                   reader.onloadend = function () {
    //                     const imageElem = new Image();
    //                     imageElem.src = reader.result;
    //                     document.getElementById('imageContainer').appendChild(imageElem);
    //                   }
    //                 })
    //             }
    //         });
    //     } catch (err) {
    //         console.log("ERROR: ", err);
    //     }
    //   }
    // }
  };
  
  const formats = [
    'bold',
    'italic',
    'strike',
    'underline',
    'list',
    'size',
    'header',
    'align',
    'indent',
    // 'font',
  ]
  
  const quill = new Quill('#editor', {
    theme: 'snow',
    formats: formats,
    modules: {
      toolbar: toolbarOptions
    }
  });

  function editorVisibility(visible) {
    document.getElementById('editor').style.display = visible ? 'block' : 'none';
    document.getElementById('toolbar').style.display = visible ? 'block' : 'none';
  };

  function previewVisibility(visible, data) {
    if (visible) {
      const imageElem = new Image();
      imageElem.src = data;
      imageElem.id = 'preview-image';
      document.getElementById('preview-container').appendChild(imageElem);
    } else {
      const img = document.getElementById('preview-image');
      if (img !== null) {
        img.remove();
      }
    }
  };

  function previewButtonVisibility(visible) {
    const previewPrintButton = document.querySelector('#preview-print-button');
    previewPrintButton.style.display = visible ? 'inline-flex' : 'none';

    const previewCancelButton = document.querySelector('#preview-cancel-button');
    previewCancelButton.style.display = visible ? 'inline-flex' : 'none';

    const previewPreviewButton = document.querySelector('#preview-preview-button');
    previewPreviewButton.style.display = visible ? 'inline-flex' : 'none';
  };

  function generateHTML() {
    var text = '<style> .text { font-family: \'Arial\'; } .footer { padding-bottom: 0.5in; } </style>';
    text = text + '<div id="text-area" class="text">' + quill.getSemanticHTML() + '</div>';
    text = text + '<div id="text-footer" class="footer"></div>';
    return text
  }

  function printHandler(previewOnly) {
    var text = generateHTML();
    console.log(text);
    console.log("Sending request to backend!");

    let url = new URL(window.location.origin + "{{ .PrintEndpoint }}");
    if (previewOnly) {
      url.searchParams.set("preview", "true");
    }
    try {
        fetch(url, {
            method: "POST",
            body: text,
        }).then((resp) => {
            if (!resp.ok) {
                console.log("error:", resp.statusText);
            } else {
              if (previewOnly) {
                resp.blob().then((blob => {
                  const url = URL.createObjectURL(blob);
                  const pdfWindow = window.open("");
                  pdfWindow.document.write("<iframe width='100%' height='100%' src='" + url + "'></iframe>");              
                }));
              }
            }
        });
    } catch (err) {
        console.log("ERROR: ", err);
    }
  }

  function onPrintButton() {
    console.log('print button pressed');
    printHandler(false);
  };

  function onCancelButton() {
    console.log('cancel button pressed');
    editorVisibility(true);
    previewVisibility(false, null);
    previewButtonVisibility(false);
  };

  function onPreviewButton() {
    console.log('preview button pressed');
    printHandler(true);
  }

  function renderPreview(data) {
    editorVisibility(false);
    previewVisibility(true, data);
    previewButtonVisibility(true);

    const previewPrintButton = document.querySelector('#preview-print-button');
    previewPrintButton.addEventListener('click', onPrintButton);

    const previewCancelButton = document.querySelector('#preview-cancel-button');
    previewCancelButton.addEventListener('click', onCancelButton);

    const previewPreviewButton = document.querySelector('#preview-preview-button');
    previewPreviewButton.addEventListener('click', onPreviewButton);
  };
  
  const printButton = document.querySelector('#print-button');
  printButton.addEventListener('click', function () {
    var text = generateHTML();
    console.log(text);
    console.log("Sending request to backend!");
    try {
        fetch("{{ .PreviewEndpoint }}", {
            method: "POST",
            body: text,
        }).then((resp) => {
            if (!resp.ok) {
                console.log("error:", resp.statusText);
            } else {
                resp.blob().then((blob) => {
                  const reader = new FileReader();
                  reader.readAsDataURL(blob);
                  reader.onloadend = function () {
                    renderPreview(reader.result);
                  }
                })
            }
        });
    } catch (err) {
        console.log("ERROR: ", err);
    }
  });