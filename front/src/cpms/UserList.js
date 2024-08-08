import React, { useState, useEffect } from 'react';
import { v4 as uuidv4 } from 'uuid';
import './UserList.css';

const adjectives = ['Happy', 'Clever', 'Brave', 'Kind', 'Witty'];
const nouns = ['Panda', 'Tiger', 'Eagle', 'Dolphin', 'Fox'];

const generateName = () => {
  const adj = adjectives[Math.floor(Math.random() * adjectives.length)];
  const noun = nouns[Math.floor(Math.random() * nouns.length)];
  return `${adj}${noun}`;
};

const UsersList = () => {
  const [users, setUsers] = useState([]);
  const [currentUser, setCurrentUser] = useState(null);
  const [ws, setWs] = useState(null);

  useEffect(() => {
    const newWs = new WebSocket('ws://localhost:8080/ws');
    const userId = uuidv4();
    const userName = generateName();

    newWs.onopen = () => {
      const user = { id: userId, name: userName };
      setCurrentUser(user);
      newWs.send(JSON.stringify({ type: 'connect', id: userId, name: userName }));
    };

    newWs.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'users') {
        setUsers(data.users);
      }
    };

    setWs(newWs);

    return () => newWs.close();
  }, []);

  return (
    <div className="users-list">
      <h3>Connected Users:</h3>
      <ul>
        {users.map((user) => (
          <li 
            key={user.id} 
            className={`user-item ${user.id === currentUser?.id ? 'current-user' : ''} ${user.isHost ? 'host' : ''}`}
          >
            {user.name}
            {user.isHost && <span className="badge host-badge">Host</span>}
            {user.id === currentUser?.id && <span className="badge current-user-badge">You</span>}
          </li>
        ))}
      </ul>
    </div>
  );
};

export default UsersList;
