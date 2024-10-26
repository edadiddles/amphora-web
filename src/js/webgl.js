const vertexShaderSource = `
    attribute vec4 aVertexPosition;
    uniform mat4 uMatrix;

    void main() {
        gl_Position = uMatrix * aVertexPosition;
        gl_PointSize = 1.0;
    }
`;

const fragmentShaderSource = `
    uniform lowp vec4 uColor;
    
    void main() {
        gl_FragColor = uColor;
    }
`;

var positions = {
    phone: [],
    paraboloid: [],
    user: [],
};

var colors = {
    phone: [1.0, 0.0, 0.0, 1.0],
    paraboloid: [0.0, 1.0, 0.0, 1.0],
    user: [0.0, 0.0, 1.0, 1.0],
}


var currTranslation = [0, 0, 0];
var currRotation = [degToRad(0), degToRad(0), degToRad(0)];
var currScale = [180,180,180];
var currDepth = 1000;
var nComponents = 3;
var shouldNormalize = true;

main();

document.querySelector("button").onclick = buttonClickHandler.bind(document);

document.querySelector("canvas").onmousedown = mouseDownHandler.bind(document);
document.querySelector("canvas").onmouseup = mouseUpHandler.bind(document);
document.querySelector("canvas").onwheel = mouseWheelHandler.bind(document);
document.querySelector("canvas").onmousemove = mouseMoveHandler.bind(document);
document.querySelector("canvas").onmouseleave = mouseLeaveHandler.bind(document);



var mouseLeftPressed = false;
var mouseWheelPressed = false;
var mouseRightPressed = false;
function mouseLeaveHandler(evt) {
    mouseLeftPressed = false;
    mouseWheelPressed = false;
    mouseRightPressed = false;
}

function mouseDownHandler(evt) {
    if(evt.which == 1) {
        mouseLeftPressed = true;
    }
    else if(evt.which == 2) {
        mouseWheelPressed = true;
    }
    else if(evt.which == 3) {
        mouseRightPressed = true;
    }
}

function mouseUpHandler(evt) {
    if(evt.which == 1) {
        mouseLeftPressed = false;
    }
    else if(evt.which == 2) {
        mouseWheelPressed = false;
    }
    else if(evt.which == 3) {
        mouseRightPressed = false;
    }
}

function mouseWheelHandler(evt) {
    var delta = -evt.deltaY/10;
    currScale[0] += delta;
    currScale[1] += delta;
    currScale[2] += delta;

    if(evt.ctrlKey) {
        currDepth += delta
        evt.preventDefault();
    }
    else if(currScale[0] < 0) {
        currScale[0] = 0;
        currScale[1] = 0;
        currScale[2] = 0;
    }
}

function mouseMoveHandler(evt) {
    if(evt.ctrlKey) {
        if(mouseLeftPressed) {
            currRotation[0] += evt.movementX/100;
        }
        else if(mouseRightPressed) {
            currRotation[1] += evt.movementY/100;
        }
        if(mouseWheelPressed) {
            currRotation[2] += evt.movementX;
        }
    }else {
        if(mouseLeftPressed) {
            currRotation[1] += evt.movementX/100;
            currRotation[0] += evt.movementY/100;
        }
        else if(mouseRightPressed) {
            currRotation[2] += evt.movementX/100;
            currRotation[0] += evt.movementY/100;
        }
        if(mouseWheelPressed) {
            currTranslation[0] += evt.movementX;
            currTranslation[1] += evt.movementY;
        }
    }
}


function buttonClickHandler() {
    //Phone
    var phone = {
        filename: document.querySelector("#phoneSelector").value,
        angle: Number(document.querySelector("#phoneAngle").value),
        angleUnits: document.getElementById("phoneAngleUnits").value,
    }

    //Paraboloid
    var paraboloid = {
        x: Number(document.querySelector("#paraboloidX").value),
        y: Number(document.querySelector("#paraboloidY").value),
        z: Number(document.querySelector("#paraboloidZ").value),
        angle: Number(document.querySelector("#paraboloidAngle").value),
        angleUnits: document.getElementById("paraboloidAngleUnits").value,
    }

    //Slicing Plane
    var slicingPlane = {
        height: Number(document.querySelector("#slicingPlaneHeight").value),
        heightUnits: document.getElementById("slicingPlaneHeightUnits").value,
        angle: Number(document.querySelector("#slicingPlaneAngle").value),
        angleUnits: document.getElementById("slicingPlaneAngleUnits").value,
    }

    //User Radius
    var userRadius = {
        radius: Number(document.querySelector("#userRadius").value),
        radiusUnits: document.getElementById("userRadiusUnits").value,
    }

    //Resolution
    var resolution = {
        linear: Number(document.querySelector("#linearResolution").value),
        angular: Number(document.querySelector("#angularResolution").value),
    }

    var payload = {
        phone: phone,
        paraboloid: paraboloid,
        slicingPlane: slicingPlane,
        userRadius: userRadius,
        resolution: resolution,
    }
    getSimulation(payload);
}


//
// start here
//
function main() {
    const canvas = document.querySelector("#glViewport");
    // Initialize the GL context
    const gl = canvas.getContext("webgl");

    // Only continue if WebGL is available and working
    if (gl === null) {
        alert("Unable to initialize WebGL. Your browser or machine may not support it.",);
        return;
    }

    currTranslation = [gl.canvas.clientWidth/2, gl.canvas.clientHeight/2, 0];

    const shaderProgram = initShaderProgram(gl, vertexShaderSource, fragmentShaderSource);

    const programInfo = {
        program: shaderProgram,
        attribLocations: {
            vertexPosition: gl.getAttribLocation(shaderProgram, 'aVertexPosition'),
        },
        uniformLocations: {
            matrix: gl.getUniformLocation(shaderProgram, 'uMatrix'),
            vertexColor: gl.getUniformLocation(shaderProgram, 'uColor'),
        },
    }

    setInterval(() => {
        // Set clear color to black, fully opaque
        gl.clearColor(0.0, 0.0, 0.0, 1.0);
        gl.clearDepth(1.0);
        gl.enable(gl.DEPTH_TEST);
        gl.enable(gl.CULL_FACE);
        // gl.enable(gl.MULTISAMPLE);
        gl.enable(gl.BLEND);
        // gl.depthRange(0.0, 10000.0);
        // gl.depthMask(false);
        gl.depthFunc(gl.LEQUAL);
        // gl.depthFunc(gl.GEQUAL);
        // gl.depthFunc(gl.EQUAL);
        
        // Clear the color buffer with specified clear color
        gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
        
        //init matrix
        var translation = currTranslation;
        var rotation = currRotation;
        var scale = currScale;

        //projection
        var matrix = m4.projection(gl.canvas.clientWidth, gl.canvas.clientHeight, currDepth);

        //translation
        matrix = m4.translate(matrix, translation[0], translation[1], translation[2]);

        //rotation
        matrix = m4.xRotate(matrix, rotation[0]);
        matrix = m4.yRotate(matrix, rotation[1]);
        matrix = m4.zRotate(matrix, rotation[2]);

        //scale
        matrix = m4.scale(matrix, scale[0], scale[1], scale[2]);

        for(k in positions) {
            var tPositions = positions[k];
            var tColor = colors[k];
            const buffers = initBuffers(gl, tPositions);
        
            // Draw the scene
            drawScene(gl, programInfo, buffers, tPositions, tColor, matrix);
        }
    }, 50);
}

function initShaderProgram(gl, vShaderSource, fShaderSource) {
    // load shaders
    const vertexShader = loadShader(gl, gl.VERTEX_SHADER, vShaderSource);
    const fragmentShader = loadShader(gl, gl.FRAGMENT_SHADER, fShaderSource);

    // create shader program
    const shaderProgram = gl.createProgram();
    // attach shaders
    gl.attachShader(shaderProgram, vertexShader);
    gl.attachShader(shaderProgram, fragmentShader);
    gl.linkProgram(shaderProgram);

    // alert program creation errors
    if (!gl.getProgramParameter(shaderProgram, gl.LINK_STATUS)) {
        alert('Unable to initialize the shader program: ' + gl.getProgramInfoLog(shaderProgram));
        return null;
    }

    return shaderProgram;
}

function loadShader(gl, type, source) {
    const shader = gl.createShader(type);
    gl.shaderSource(shader, source);
    gl.compileShader(shader);

    if(!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
        alert('An error occurred compiling the shaders: ' + gl.getShaderInfoLog(shader));
        gl.deleteShader(shader);
        return null;
    }

    return shader;
}

function initBuffers(gl, verticies) {
    const positionBuffer = initPositionBuffer(gl, verticies);

    return {
        position: positionBuffer,
    };
}

function initPositionBuffer(gl, verticies) {
    const positionBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(verticies), gl.STATIC_DRAW);

    return positionBuffer;
}

function drawScene(gl, programInfo, buffers, positions, color, matrix) {
    {
        const numComponents = nComponents;
        const type = gl.FLOAT;
        const normalize = shouldNormalize;
        const stride = 0;
        const offset = 0;
        gl.bindBuffer(gl.ARRAY_BUFFER, buffers.position);
        gl.vertexAttribPointer(
            programInfo.attribLocations.vertexPosition,
            numComponents,
            type,
            normalize,
            stride,
            offset,
        );
        gl.enableVertexAttribArray(programInfo.attribLocations.vertexPosition);
    }
    gl.useProgram(programInfo.program);
    gl.uniform4fv(programInfo.uniformLocations.vertexColor, color);
   
    gl.uniformMatrix4fv(programInfo.uniformLocations.matrix, false, matrix);

    {
        const offset = 0;
        const vertexCount = positions.length/3;
        gl.drawArrays(gl.POINTS, offset, vertexCount);
    }
}

function getSimulation(payload) {
    opts = {
        method: "POST",
        body: JSON.stringify(payload),
    }
    fetch(`http://localhost:8080/api/simulation`, opts).then(function(response) {
        return response.json();
    }).then(function(data) {
        positions.phone = data.Phone || [];
        positions.paraboloid = data.Paraboloid || [];
        positions.user = data.User || [];
    });
}




function degToRad(deg) {
    return deg*Math.PI/180;
}

var m4 = {

  projection: function(width, height, depth) {
    // Note: This matrix flips the Y axis so 0 is at the top.
    return [
       2 / width, 0, 0, 0,
       0, -2 / height, 0, 0,
       0, 0, 2 / depth, 0,
      -1, 1, 0, 1,
    ];
  },

  multiply: function(a, b) {
    var a00 = a[0 * 4 + 0];
    var a01 = a[0 * 4 + 1];
    var a02 = a[0 * 4 + 2];
    var a03 = a[0 * 4 + 3];
    var a10 = a[1 * 4 + 0];
    var a11 = a[1 * 4 + 1];
    var a12 = a[1 * 4 + 2];
    var a13 = a[1 * 4 + 3];
    var a20 = a[2 * 4 + 0];
    var a21 = a[2 * 4 + 1];
    var a22 = a[2 * 4 + 2];
    var a23 = a[2 * 4 + 3];
    var a30 = a[3 * 4 + 0];
    var a31 = a[3 * 4 + 1];
    var a32 = a[3 * 4 + 2];
    var a33 = a[3 * 4 + 3];
    var b00 = b[0 * 4 + 0];
    var b01 = b[0 * 4 + 1];
    var b02 = b[0 * 4 + 2];
    var b03 = b[0 * 4 + 3];
    var b10 = b[1 * 4 + 0];
    var b11 = b[1 * 4 + 1];
    var b12 = b[1 * 4 + 2];
    var b13 = b[1 * 4 + 3];
    var b20 = b[2 * 4 + 0];
    var b21 = b[2 * 4 + 1];
    var b22 = b[2 * 4 + 2];
    var b23 = b[2 * 4 + 3];
    var b30 = b[3 * 4 + 0];
    var b31 = b[3 * 4 + 1];
    var b32 = b[3 * 4 + 2];
    var b33 = b[3 * 4 + 3];
    return [
      b00 * a00 + b01 * a10 + b02 * a20 + b03 * a30,
      b00 * a01 + b01 * a11 + b02 * a21 + b03 * a31,
      b00 * a02 + b01 * a12 + b02 * a22 + b03 * a32,
      b00 * a03 + b01 * a13 + b02 * a23 + b03 * a33,
      b10 * a00 + b11 * a10 + b12 * a20 + b13 * a30,
      b10 * a01 + b11 * a11 + b12 * a21 + b13 * a31,
      b10 * a02 + b11 * a12 + b12 * a22 + b13 * a32,
      b10 * a03 + b11 * a13 + b12 * a23 + b13 * a33,
      b20 * a00 + b21 * a10 + b22 * a20 + b23 * a30,
      b20 * a01 + b21 * a11 + b22 * a21 + b23 * a31,
      b20 * a02 + b21 * a12 + b22 * a22 + b23 * a32,
      b20 * a03 + b21 * a13 + b22 * a23 + b23 * a33,
      b30 * a00 + b31 * a10 + b32 * a20 + b33 * a30,
      b30 * a01 + b31 * a11 + b32 * a21 + b33 * a31,
      b30 * a02 + b31 * a12 + b32 * a22 + b33 * a32,
      b30 * a03 + b31 * a13 + b32 * a23 + b33 * a33,
    ];
  },

  translation: function(tx, ty, tz) {
    return [
       1,  0,  0,  0,
       0,  1,  0,  0,
       0,  0,  1,  0,
       tx, ty, tz, 1,
    ];
  },

  xRotation: function(angleInRadians) {
    var c = Math.cos(angleInRadians);
    var s = Math.sin(angleInRadians);

    return [
      1, 0, 0, 0,
      0, c, s, 0,
      0, -s, c, 0,
      0, 0, 0, 1,
    ];
  },

  yRotation: function(angleInRadians) {
    var c = Math.cos(angleInRadians);
    var s = Math.sin(angleInRadians);

    return [
      c, 0, -s, 0,
      0, 1, 0, 0,
      s, 0, c, 0,
      0, 0, 0, 1,
    ];
  },

  zRotation: function(angleInRadians) {
    var c = Math.cos(angleInRadians);
    var s = Math.sin(angleInRadians);

    return [
       c, s, 0, 0,
      -s, c, 0, 0,
       0, 0, 1, 0,
       0, 0, 0, 1,
    ];
  },

  scaling: function(sx, sy, sz) {
    return [
      sx, 0,  0,  0,
      0, sy,  0,  0,
      0,  0, sz,  0,
      0,  0,  0,  1,
    ];
  },

  translate: function(m, tx, ty, tz) {
    return m4.multiply(m, m4.translation(tx, ty, tz));
  },

  xRotate: function(m, angleInRadians) {
    return m4.multiply(m, m4.xRotation(angleInRadians));
  },

  yRotate: function(m, angleInRadians) {
    return m4.multiply(m, m4.yRotation(angleInRadians));
  },

  zRotate: function(m, angleInRadians) {
    return m4.multiply(m, m4.zRotation(angleInRadians));
  },

  scale: function(m, sx, sy, sz) {
    return m4.multiply(m, m4.scaling(sx, sy, sz));
  },

};
