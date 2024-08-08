import './App.css';
import React, { useRef, useEffect, useState } from 'react';
import DrawingCanvas from './cpms/DrawingCanvas';
import UsersList from './cpms/UserList';

function App() {
  const [currentUser, setCurrentUser] = useState(null);
  const [isHost, setIsHost] = useState(false);

  return (
    <div className="App">
      <div className="main-content">
        <DrawingCanvas isHost={isHost} currentUser={currentUser} />
      </div>
      <div className="sidebar">
        <UsersList setCurrentUser={setCurrentUser} setIsHost={setIsHost} />
      </div>
    </div>
  );
}

export default App;
