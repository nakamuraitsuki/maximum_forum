import "./App.css";
import { Route, Routes, Link } from "react-router-dom";
import { useState, useEffect } from "react";
import Login from "./pages/Login";
import Register from "./pages/Register";

function App() {
  const [message, setMessage] = useState("");
  const [comments, setComments] = useState([]);
  const [getTrigger, setGetTrigger] = useState(false);

  const getComments = async () => {
    const url = "http://localhost:8080/api/comments";
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`コメント取得エラー/status:${response.status}`);
      }
      const data = await response.json();
      console.log("コメント取得成功", data);
      if (data != null) setComments(data);
    } catch (error) {
      console.error(error.message);
    }
  };

  const postMessage = async (message) => {
    try {
      const response = await fetch("http://localhost:8080/api/comments", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ message: message }),
      });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      console.log("コメントが投稿されました:", data);
      setGetTrigger((prev) => !prev);
      return data;
    } catch (error) {
      console.error("Fetchエラーが発生しました:", error);
    }
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    postMessage(message);
    setMessage("");
  };

  useEffect(() => {
    getComments();
  }, [getTrigger]);

  return (
    <div className="App">
      <h1>Maximum掲示板</h1>
      <nav>
        <Link to="/login">ログイン</Link>
        <Link to="/register">新規登録</Link>
      </nav>
      
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
      </Routes>

      <div>
        {comments.map((comment) => (
          <div key={comment.id}>
            <p>{comment.message}</p>
          </div>
        ))}
      </div>

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
