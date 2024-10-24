import './App.css';
import { useState } from 'react'

function App() {
  const [message, setMessage] = useState('');

  // メッセージを投稿する関数
  const postMessage = async (message) => {
    try {
      const response = await fetch('http://localhost:8080/api/comments', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message: message }),
      });
  
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
  
      const data = await response.json();
      console.log('コメントが投稿されました:', data);
      return data;
    } catch (error) {
      console.error('Fetchエラーが発生しました:', error);
    }
  };  

  // フォーム送信時の処理
  const handleSubmit = (event) => {
    event.preventDefault(); // ページのリロードを防ぐ
    postMessage(message); // メッセージを投稿
    setMessage(''); // メッセージをリセット
  };
  return (
    <div className="App">
      <h1>Maximum掲示板</h1>
      <form onSubmit={handleSubmit}>
        <textarea
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="コメントを入力してください"
        ></textarea>
        <button type="submit">投稿</button>
      </form>
    </div>
  );
}

export default App;