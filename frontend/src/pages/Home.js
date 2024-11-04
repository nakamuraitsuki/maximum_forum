import { Link } from "react-router-dom";
import { useState, useEffect } from "react";

function Home() {
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

  return (
    <div className="App">
      <h1>Maximum掲示板</h1>
      {loggedInUser && <p>{loggedInUser} さん、こんにちは！</p>}
      <nav>
        <Link to="/register">新規登録</Link>
        {loggedInUser ? (
          <Link to="/logout">ログアウト</Link>
        ) : (
          <Link to="/login">ログイン</Link>
        )}
      </nav>
    </div>
  );
}

export default Home;
