import "./Register.css";
import { Link } from "react-router-dom";
import { useState } from "react";

export default function Register() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const postUser = (username, password) => {
    fetch("http://localhost:8080/api/users", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name: username,
        pw_hash: password,
      }),
    })
      .then((response) => {
        if (response.ok) {
          // JSONレスポンスが存在するか確認
          return response
            .json()
            .then((data) => {
              console.log("ユーザー登録成功:", data);
              window.location.href = "/login";
            })
            .catch(() => {
              console.log("ユーザー登録成功（空のレスポンス）");
              window.location.href = "/login";
            });
        } else {
          // エラーステータスに応じてエラーメッセージを設定
          if (response.status === 409) {
            setError("ユーザー名が既に使用されています");
          } else {
            setError("ユーザー登録エラーが発生しました");
          }
        }
      })
      .catch((error) => {
        console.error("登録中にエラーが発生しました:", error);
        setError("サーバーエラーが発生しました");
      });
  };

  // ユーザー名とパスワードの文字数を確認する関数
  const checkForms = () => {
    if (username.length < 3 || username.length > 20) {
      setError("ユーザー名は3文字以上20文字以下である必要があります");
      return 0;
    }
    if (password.length < 8 || password.length > 16) {
      setError("パスワードは8文字以上16文字以下である必要があります");
      return 0;
    } else if (!password.match(/^[0-9a-zA-Z]+$/)) {
      setError("パスワードは半角英数字である必要があります");
      return 0;
    }
    return 1;
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    // checkForms()が0を返す場合は、処理を中断する
    if (checkForms() === 0) {
      return;
    }
    postUser(username, password);
    setUsername("");
    setPassword("");
  };

  return (
    <div className="register-container">
      <div className="register">
        <h1>新規登録</h1>
        <nav className="register-nav">
          <Link to="/">スレッド一覧</Link>
          <Link to="/login">ログイン</Link>
        </nav>
        <form onSubmit={handleSubmit} className="register-form">
          {error && <p style={{ color: "red" }}>{error}</p>}
          <input
            type="text"
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <button type="submit">登録</button>
        </form>
      </div>
    </div>
  );
}
