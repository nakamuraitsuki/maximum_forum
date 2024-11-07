import { Link } from "react-router-dom";
import { useState, useEffect } from "react";

function Home() {
  const [threadName, setThreadName] = useState("");
  const [threads, setThreads] = useState([]);
  const [getTrigger, setGetTrigger] = useState(false);
  const [loggedInUser, setLoggedInUser] = useState({id: "", name: ""});

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
      const id = getUserIdFromToken(token);
      const name = getUsernameFromToken(token);
      setLoggedInUser({id, name});
    }
  }, []);

  const getThreads = async () => {
    const url = "http://localhost:8080/api/threads";
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`スレッド取得エラー/status:${response.status}`);
      }
      const data = await response.json();
      console.log("スレッド取得成功", data);
      if (data != null) setThreads(data);
    } catch (error) {
      console.error(error.message);
    }
  };

  const postThread = async (threadName) => {
    const token = document.cookie.replace(
      /(?:(?:^|.*;\s*)token\s*=\s*([^;]*).*$)|^.*$/,
      "$1"
    ); // トークンを取得

    if (!token) {
      console.error("トークンがありません。ログインが必要です。");
      return;
    }

    try {
      const response = await fetch("http://localhost:8080/api/threads", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`, // 認証用のトークンを追加
        },
        body: JSON.stringify({
          name: threadName
        }),
      });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      console.log("スレッド作成:", data);
      setGetTrigger((prev) => !prev);
      return data;
    } catch (error) {
      console.error("Fetchエラーが発生しました:", error);
    }
  };

  const deleteThread = async (threadID) => {
    try{
      const response = await fetch(`http://localhost:8080/api/threads/${threadID}`, {
        method: "DELETE",
      });
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      console.log("スレッド削除:");
      setGetTrigger((prev) => !prev);
      return;
    } catch (error) {
      console.log("Fetch Error", error);
    }
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    postThread(threadName);
    setThreadName("");
  };

  useEffect(() => {
    getThreads();
  }, [getTrigger]);

  return (
    <div className="App">
      <h1>Maximum掲示板</h1>
      {loggedInUser && <p>{loggedInUser.name} さん、こんにちは！</p>}
      <nav>
        <Link to="/register">新規登録</Link>
        {loggedInUser ? (
          <Link to="/logout">ログアウト</Link>
        ) : (
          <Link to="/login">ログイン</Link>
        )}
      </nav>
      <div>
        {threads.map((thread) => (
          <div key={thread.id}>
            <Link to={`/thread/${thread.id}`}>
              <span>
                {thread.name}{" "}
                {new Date(thread.created_at).toLocaleString()}
              </span>
            </Link>
            {loggedInUser.id == thread.owner_id &&         <button type="button">削除</button> }
          </div>
        ))}
      </div>
      <form onSubmit={handleSubmit}>
        <textarea
          value={threadName}
          onChange={(e) => setThreadName(e.target.value)}
          placeholder="スレッド名"
          required
        ></textarea>
        <button type="submit">作成</button>
      </form>
    </div>
  );
}

export default Home;
