import "./Home.css";
import { Link } from "react-router-dom";
import { useState, useEffect, useRef } from "react";

function Home() {
  const [threadName, setThreadName] = useState("");
  const [threads, setThreads] = useState([]);
  const [getTrigger, setGetTrigger] = useState(false);
  const [loggedInUser, setLoggedInUser] = useState({ id: "", name: "" });
  const searchInputRef = useRef();

  const getUsernameFromToken = (token) => {
    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      return payload.name;
    } catch (error) {
      console.error("トークンからユーザー名を取得できませんでした:", error);
      return "";
    }
  };

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
    const token = document.cookie.replace(
      /(?:(?:^|.*;\s*)token\s*=\s*([^;]*).*$)|^.*$/,
      "$1"
    );

    if (token) {
      const id = getUserIdFromToken(token);
      const name = getUsernameFromToken(token);
      setLoggedInUser({ id, name });
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
      else setThreads([]);
    } catch (error) {
      console.error(error.message);
    }
  };

  const postThread = async (threadName) => {
    const token = document.cookie.replace(
      /(?:(?:^|.*;\s*)token\s*=\s*([^;]*).*$)|^.*$/,
      "$1"
    );

    if (!token) {
      console.error("トークンがありません。ログインが必要です。");
      return;
    }

    try {
      const response = await fetch("http://localhost:8080/api/threads", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ name: threadName }),
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
    try {
      const response = await fetch(
        `http://localhost:8080/api/threads/${threadID}`,
        {
          method: "DELETE",
        }
      );
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      console.log("スレッド削除:");
      setGetTrigger((prev) => !prev);
    } catch (error) {
      console.error("Fetch Error", error);
    }
  };

  const threadFilter = (event) => {
    event.preventDefault();
    const keyword = event.target[0].value;
    if (keyword === "") {
      setGetTrigger((prev) => !prev);
      return;
    }
    const filteredThreads = threads.filter((thread) =>
      thread.name.includes(keyword)
    );
    if (filteredThreads.length === 0) {
      alert("スレッドが見つかりませんでした");
    } else {
      setThreads(filteredThreads);
    }
  };

  const handleReset = () => {
    setGetTrigger((prev) => !prev);
    if (searchInputRef.current) {
      searchInputRef.current.value = "";
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
    <div className="home">
      <h1>Maximum掲示板</h1>
      {loggedInUser.id && <p>{loggedInUser.name} さん、こんにちは！</p>}
      <nav className="home-nav">
        <Link to="/register">新規登録</Link>
        {loggedInUser.id && loggedInUser.name ? (
          <Link to="/logout">ログアウト</Link>
        ) : (
          <Link to="/login">ログイン</Link>
        )}
      </nav>
      <div className="thread-filter">
        <form onSubmit={threadFilter}>
          <input type="text" placeholder="スレッド検索" ref={searchInputRef} />
          <button type="submit">検索</button>
          <button type="button" onClick={handleReset}>
            リセット
          </button>
        </form>
      </div>
      <div className="comments">
        {threads.map((thread) => (
          <div key={thread.id}>
            <Link to={`/thread/${thread.id}`}>
              <span>
                {thread.name} {new Date(thread.created_at).toLocaleString()}
              </span>
            </Link>
            {loggedInUser.id === String(thread.owner_id) && (
              <button type="button" onClick={() => deleteThread(thread.id)}>
                削除
              </button>
            )}
          </div>
        ))}
      </div>
      <form onSubmit={handleSubmit} className="comment-form">
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
