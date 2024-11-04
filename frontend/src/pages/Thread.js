import { Link } from "react-router-dom";
import { useState, useEffect } from "react";
import { useParams } from "react-router-dom";

function Thread() {
  const { thread_id } = useParams();
  const [message, setMessage] = useState("");
  const [comments, setComments] = useState([]);
  const [threadInfo, setThreadInfo] = useState([]);
  const [getTrigger, setGetTrigger] = useState(false);
  const [loggedInUser, setLoggedInUser] = useState("");

  // JWTトークンからユーザー名を取得する関数
  const getUsernameFromToken = (token) => {
    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      return payload.name;
    } catch (error) {
      console.error("トークンからユーザー名を取得できませんでした:", error);
      return "";
    }
  };

  // JWTトークンからユーザーIDを取得する関数
  const getUserIdFromToken = (token) => {
    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      return payload.user_id;
    } catch (error) {
      console.error("トークンからユーザーIDを取得できませんでした:", error);
      return "";
    }
  };

  useEffect(() => {
    // トークンが存在する場合、ユーザー名を取得してセット
    const token = document.cookie.replace(
      /(?:(?:^|.*;\s*)token\s*=\s*([^;]*).*$)|^.*$/,
      "$1"
    );

    if (token) {
      const name = getUsernameFromToken(token);
      setLoggedInUser(name);
    }
  }, []);

  const getComments = async () => {
    const url = `http://localhost:8080/api/comments?threadID=${thread_id}`;
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

  const getThreadInfo = async () => {
    const url = `http://localhost:8080/api/threads?threadID=${thread_id}`
    try {
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error(`コメント取得エラー/status:${response.status}`);
        }
        const data = await response.json();
        console.log("コメント取得成功", data);
        if (data != null) setThreadInfo(data);
      } catch (error) {
        console.error(error.message);
      }
  }

  const postMessage = async (message) => {
    const token = document.cookie.replace(
      /(?:(?:^|.*;\s*)token\s*=\s*([^;]*).*$)|^.*$/,
      "$1"
    ); // トークンを取得
    const userId = getUserIdFromToken(token); // ユーザーIDを取得

    if (!token) {
      console.error("トークンがありません。ログインが必要です。");
      return;
    }

    try {
      const response = await fetch("http://localhost:8080/api/comments", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`, // 認証用のトークンを追加
        },
        body: JSON.stringify({
          user_id: userId,
          thread_id: Number(thread_id),
          message: message,
        }),
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
      <h1>Maximum掲示板:スレッド{threadInfo.id}</h1>
      {loggedInUser && <p>{loggedInUser} さん、こんにちは！</p>}
      <nav>
        <Link to="/register">新規登録</Link>
        {loggedInUser ? (
          <Link to="/logout">ログアウト</Link>
        ) : (
          <Link to="/login">ログイン</Link>
        )}
      </nav>

      <div>
        {comments.map((comment) => (
          <div key={comment.id}>
            <p>
              {comment.name}:{comment.message}{" "}
              {new Date(comment.created_at).toLocaleString()}
            </p>
          </div>
        ))}
      </div>

      <form onSubmit={handleSubmit}>
        <textarea
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          placeholder="コメントを入力してください"
          required
        ></textarea>
        <button type="submit">投稿</button>
      </form>
    </div>
  );
}

export default Thread;
