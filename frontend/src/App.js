import './App.css';

function App() {
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

  return (
    <div className="App">
      <h1>Maximum掲示板</h1>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          const message = e.target.elements.message.value;
          e.target.elements.message.value = '';
          await postMessage(message);
        }}
      >
        <input type="text" required name="message" />
        <button type="submit">投稿</button>
      </form>
    </div>
  );
}

export default App;