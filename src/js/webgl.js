const vertexShaderSource = `
    attribute vec4 aVertexPosition;
    // uniform mat4 uMatrix;

    void main() {
        gl_Position = aVertexPosition;
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

main();

document.querySelector("button").onclick = buttonClickHandler.bind(document);

function buttonClickHandler() {
    //Phone
    var phone = document.querySelector("#phoneSelector").value

    //Paraboloid
    var paraboloid = {
        x: Number(document.querySelector("#paraboloidX").value),
        y: Number(document.querySelector("#paraboloidY").value),
        z: Number(document.querySelector("#paraboloidZ").value),
        angle: Number(document.querySelector("#paraboloidAngle").value),
    }

    //Slicing Plane
    var slicingPlane = {
        height: Number(document.querySelector("#slicingPlaneHeight").value),
    }

    //User Radius
    var userRadius = {
        radius: Number(document.querySelector("#userRadius").value),
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

    const shaderProgram = initShaderProgram(gl, vertexShaderSource, fragmentShaderSource);

    const programInfo = {
        program: shaderProgram,
        attribLocations: {
            vertexPosition: gl.getAttribLocation(shaderProgram, 'aVertexPosition'),
        },
        uniformLocations: {
            projectionMatrix: gl.getUniformLocation(shaderProgram, 'uProjectionMatrix'),
            modelViewMatrix: gl.getUniformLocation(shaderProgram, 'uModelViewMatrix'),
            vertexColor: gl.getUniformLocation(shaderProgram, 'uColor')
        },
    }

    setInterval(() => {
        // Set clear color to black, fully opaque
        gl.clearColor(0.0, 0.0, 0.0, 1.0);
        gl.clearDepth(1.0);
        gl.enable(gl.DEPTH_TEST);
        gl.depthFunc(gl.LEQUAL);
        
        // Clear the color buffer with specified clear color
        gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
        
        for(k in positions) {
            var tPositions = positions[k];
            var tColor = colors[k];
            const buffers = initBuffers(gl, tPositions);
        
            // Draw the scene
            drawScene(gl, programInfo, buffers, tPositions, tColor);
        }
    }, 1000);
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

function drawScene(gl, programInfo, buffers, positions, color) {
    {
        const numComponents = 3;
        const type = gl.FLOAT;
        const normalize = false;
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
    
    {
        const offset = 0;
        const vertexCount = positions.length;
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
        positions.phone = data.Phone;
        positions.paraboloid = data.Paraboloid;
        positions.user = data.User;
    });
}