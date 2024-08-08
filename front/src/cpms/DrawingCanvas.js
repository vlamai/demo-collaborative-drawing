import React, { useRef, useEffect, useState } from 'react';
import './DrawingCanvas.css';

const DrawingCanvas = ({ isHost, currentUser }) => {
  const canvasRef = useRef(null);
  const [canvasSize, setCanvasSize] = useState({ width: 0, height: 0 });
  const [ws, setWs] = useState(null);

  useEffect(() => {
    const newWs = new WebSocket('ws://localhost:8080/ws');

    newWs.onopen = () => {
      if (isHost) {
        updateCanvasSize();
        window.addEventListener('resize', updateCanvasSize);
      }
    };

    newWs.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'canvasSize' && !isHost) {
        setCanvasSize(data.size);
      }
    };

    setWs(newWs);

    return () => {
      newWs.close();
      if (isHost) {
        window.removeEventListener('resize', updateCanvasSize);
      }
    };
  }, [isHost]);

  const updateCanvasSize = () => {
    if (isHost && canvasRef.current) {
      const newSize = {
        width: canvasRef.current.offsetWidth,
        height: canvasRef.current.offsetHeight,
      };
      setCanvasSize(newSize);
      ws.send(JSON.stringify({ type: 'canvasSize', size: newSize }));
    }
  };

  useEffect(() => {
    if (canvasRef.current) {
      const canvas = canvasRef.current;
      const ctx = canvas.getContext('2d');

      if (isHost) {
        canvas.width = canvasSize.width;
        canvas.height = canvasSize.height;
      } else {
        const containerWidth = canvas.parentElement.clientWidth;
        const containerHeight = canvas.parentElement.clientHeight;
        const aspectRatio = canvasSize.width / canvasSize.height;

        let width = containerWidth;
        let height = containerWidth / aspectRatio;

        if (height > containerHeight) {
          height = containerHeight;
          width = containerHeight * aspectRatio;
        }

        canvas.width = width;
        canvas.height = height;
        canvas.style.width = `${width}px`;
        canvas.style.height = `${height}px`;
      }

      // Clear the canvas
      ctx.fillStyle = 'white';
      ctx.fillRect(0, 0, canvas.width, canvas.height);
    }
  }, [canvasSize, isHost]);

  return (
    <div className={`canvas-container ${isHost ? 'host' : 'guest'}`}>
      <canvas ref={canvasRef} />
    </div>
  );
};

export default DrawingCanvas;
