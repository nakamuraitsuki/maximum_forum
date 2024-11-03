import { useState } from "react";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  const postLogin = (username, password) => {
    fetch("http://localhost:8080/api/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name: username,
        pw_hash: password,
      }),
    }).then((response) => {
      if (response.ok) {
        response.json().then((data) => {
          console.log("ログイン成功:", data);
          localStorage.setItem("token", data.token);
          window.location.href = "/";
        });
      } else {
        if(response.status === 401) {
          setError("ユーザー名またはパスワードが間違っています");
        } else {
          setError("ログインエラーが発生しました");
        }
      }
    });
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    postLogin(username, password);
    setUsername("");
    setPassword("");
  };

  return (
    <div>
      <h1>Login</h1>
      <form onSubmit={handleSubmit}>
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
        <button type="submit">Login</button>
      </form>
    </div>
  );
}
