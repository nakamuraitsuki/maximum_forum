import './App.css';

function App() {
  const [message, setMessage] = useState('');
  const [submitted, setSubmitted] = useState(false);

  // フォーム送信時の処理
  const handleSubmit = (event) => {
    event.preventDefault(); // ページのリロードを防ぐ
    setSubmitted(true); // 投稿が完了したことを設定
  };
  return (
    <div className="App">
      <h1>Maximum掲示板</h1>
      <form onSubmit={handleSubmit}>
        <label htmlFor="message">メッセージ:</label>
        <input
          type="text"//入力
          id="message"
          name="message"
          value={message} 
          onChange={(e) => setMessage(e.target.value)} // 入力が変更されたときにstateを更新
          required
        /><br /><br />

        <button type="submit">投稿</button> 
      </form>

      {submitted && (
        <div>
          <h2>投稿が完了しました</h2>
          <p><strong>匿名</strong> さんのメッセージ:</p>
          <p>{message}</p>
        </div>
      )}
    </div>//<div>終わり
  );
}

export default App;