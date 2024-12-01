import "./Thread.css";
import { Link } from "react-router-dom";
import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";

function Thread() {
  const { thread_id } = useParams();
  const [message, setMessage] = useState("");
  const [comments, setComments] = useState([]);
  const [isLimitReached, setIsLimitReached] = useState(false);
  const [commentCount, setCommentCount] = useState({
    commentCount: 0,
    maxComment: 0,
  });
  const [threadInfo, setThreadInfo] = useState([]);
  const [getTrigger, setGetTrigger] = useState(false);
  const [loggedInUser, setLoggedInUser] = useState("");

  const navigate = useNavigate();

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
      if (data.comments != null) setComments(data.comments);
      setIsLimitReached(data.is_limit_reached);
      setCommentCount({
        commentCount: data.comment_count,
        maxComment: data.max_comments,
      });
    } catch (error) {
      console.error(error.message);
    }
  };

  const getThreadInfo = async () => {
    const url = `http://localhost:8080/api/threads/${thread_id}`;
    try {
      const response = await fetch(url);
      if (!response.ok) {
        if (response.status == 404) {
          navigate("/NotFound");
        }
        throw new Error(`スレッド取得エラー/status:${response.status}`);
      }
      const data = await response.json();
      console.log("スレッド取得成功", data);
      if (data != null) setThreadInfo(data);
    } catch (error) {
      console.error(error.message);
    }
  };

  const postMessage = async (message) => {
    const token = document.cookie.replace(
      /(?:(?:^|.*;\s*)token\s*=\s*([^;]*).*$)|^.*$/,
      "$1"
    ); // トークンを取得

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
          thread_id: Number(thread_id),
          message: message,
        }),
      });
      //コメント上限に達している場合
      if (response.status === 403) {
        console.error("コメントの上限に達しました。");
        //TODO:上限を迎えていてなおコメントをポストした際の表示
        return;
      }
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
    getThreadInfo();
  }, [getTrigger]);

  return (
    <div className="container">
      <Link to="/" className="home-link">
        <h1>Maximum掲示板</h1>
      </Link>
      <nav className="navigation">
        <Link to="/register">新規登録</Link>
        {loggedInUser ? (
          <Link to="/logout">ログアウト</Link>
        ) : (
          <Link to="/login">ログイン</Link>
        )}
        <Link to="/">スレッド一覧</Link>
      </nav>
      <h2 className="thread-title">{threadInfo.name}</h2>
      {loggedInUser && <span>{loggedInUser} さん、こんにちは！</span>}
      <div>
        <form onSubmit={handleSubmit} className="comment-form">
          <textarea
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            placeholder="コメントを入力してください"
            required
          ></textarea>
          <button type="submit">投稿</button>
        </form>
        <span>
          コメント数:{commentCount.commentCount}/{commentCount.maxComment}
        </span>
        {isLimitReached && (
          <span className="comment-limited">コメント上限に達しています</span>
        )}
        {comments.map((comment) => (
          <div key={comment.id} className="comment">
            <span className="id">
              {comments.length - comments.indexOf(comment)}
              {"."}
            </span>
            <span className="name">{comment.name}</span>
            <p className="created-at">
              {new Date(comment.created_at).toLocaleString()}
            </p>
            <p className="message">{comment.message}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

export default Thread;
