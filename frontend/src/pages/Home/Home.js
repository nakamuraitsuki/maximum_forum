import "./Home.css";
import { Link } from "react-router-dom";
import { useState, useEffect, useRef, useMemo } from "react";
import usePagination from "@mui/material/usePagination";
import { MdClear } from "react-icons/md";
import { MdSearch } from "react-icons/md";
import { MdArrowBackIos } from "react-icons/md";
import { MdArrowForwardIos } from "react-icons/md";

function Home() {
  const [threadName, setThreadName] = useState("");
  const [threads, setThreads] = useState([]);
  const [page, setPage] = useState(1);
  //現在のスレッド関連情報（スレッド上限、現在のスレッド数、現在の総ページ数）
  const [threadsInfo, setThreadsInfo] = useState({
    MaxThreads: 0,
    ThreadCount: 0,
    PageCount: 0,
    MaxComments: 0,
  });
  const [isLimitReached, setIsLimitReached] = useState(false);
  const [getTrigger, setGetTrigger] = useState(false);
  const [loggedInUser, setLoggedInUser] = useState({ id: "", name: "" });
  const [searchKeyword, setSearchKeyword] = useState("");
  const searchInputRef = useRef();
  const allThreads = useMemo(() => threads, [threads]);

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

  const getThreads = async (page) => {
    try {
      const response = await fetch(
        `http://localhost:8080/api/threads?page=${page}`
      );
      if (!response.ok) {
        throw new Error(`スレッド取得エラー/status:${response.status}`);
      }
      const data = await response.json();
      console.log("スレッド取得成功", data);
      if (data.threads != null) setThreads(data.threads);
      else setThreads([]);
      setIsLimitReached(data.is_limit_reached);
      setThreadsInfo({
        MaxThread: Number(data.max_threads),
        ThreadCount: Number(data.thread_count),
        PageCount: Number(data.page_count),
        MaxComments: Number(data.max_comments),
      });
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
      //スレッド条件に達している場合
      if (response.status === 403) {
        console.error("スレッドの上限に達しました。");
        //TODO:上限を迎えていてなおスレッドの作成をした際の表示
        return;
      }
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
        if (response.status === 404) {
          console.log(`Thread with ID ${threadID} not found.`);
          return;
        } else if (response.status === 500) {
          console.log("Server error occurred while deleting the thread.");
          return;
        } else {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
      }
      console.log("スレッド削除:");
      setGetTrigger((prev) => !prev);
    } catch (error) {
      console.error("Fetch Error", error);
    }
  };

  useEffect(() => {
    getThreads(page);
  }, [getTrigger]);

  const filteredThreads = useMemo(() => {
    return allThreads.filter((thread) =>
      thread.name.toLowerCase().includes(searchKeyword.toLowerCase())
    );
  }, [allThreads, searchKeyword]);

  const handleSearch = (event) => {
    event.preventDefault();
    const keyword = searchInputRef.current.value;
    setSearchKeyword(keyword);
  };

  const handleReset = () => {
    setSearchKeyword("");
    if (searchInputRef.current) {
      searchInputRef.current.value = "";
    }
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    postThread(threadName);
    setThreadName("");
  };
  //ページ遷移関数
  const handlePageChange = (e, newPage) => {
    getThreads(newPage);
    setPage(newPage);
  };
  //ページネーション
  const { items } = usePagination({
    count: threadsInfo.PageCount, //総ページ数
    page: page, //現在いるページ
    onChange: handlePageChange, //ページ遷移関数
    siblingCount: 1,
    boundaryCount: 1,
  });
  //表示ラベル
  const getLabel = (type, page) => {
    switch (type) {
      case "start-ellipsis":
      case "end-ellipsis":
        return "...";
      case "previous":
        return <MdArrowBackIos />;
      case "next":
        return <MdArrowForwardIos />;
      default:
        return page;
    }
  };

  return (
    <div className="home">
      <Link to="/" className="home-link">
        Maximum掲示板
      </Link>
      <img src="/images/maximum-logo.png" alt="maximum-logo" className="logo" />
      {loggedInUser.id && <p>{loggedInUser.name} さん、こんにちは！</p>}
      <nav className="home-nav">
        <Link to="/register">新規登録</Link>
        {loggedInUser.id && loggedInUser.name ? (
          <Link to="/logout">ログアウト</Link>
        ) : (
          <Link to="/login">ログイン</Link>
        )}
      </nav>
      <div className="create-thread-form">
        <form onSubmit={handleSubmit}>
          <input
            value={threadName}
            onChange={(e) => setThreadName(e.target.value)}
            placeholder="スレッド名"
            required
          ></input>
          <button type="submit">作成</button>
        </form>
      </div>
      <div className="thread-limited">
        {isLimitReached && <span>スレッド数の上限に達しています</span>}
      </div>
      <div className="thread-filter">
        <form onSubmit={handleSearch}>
          <input type="text" placeholder="スレッド検索" ref={searchInputRef} />
          <button type="button" onClick={handleReset}>
            <MdClear />
          </button>
          <button type="submit" className="submit">
            <MdSearch />
          </button>
        </form>
      </div>
      {filteredThreads.length === 0 ? (
        <p>スレッドがありません</p>
      ) : (
        <div className="thread-list">
          {filteredThreads.map((thread) => (
            <div key={thread.id} className="thread-container">
              <Link to={`/thread/${thread.id}`}>
                <span className="thread-name">{thread.name}</span>
                <span className="thread-info">
                  コメント数:{thread.comment_count}/{threadsInfo.MaxComments}・
                  {new Date(thread.created_at).toLocaleString()}
                </span>
              </Link>
              {loggedInUser.id == String(thread.owner_id) ? (
                <button
                  type="button"
                  onClick={() => deleteThread(thread.id)}
                  className="delete-button"
                >
                  削除
                </button>
              ) : (
                <button
                  type="button"
                  onClick={() => deleteThread(thread.id)}
                  className="delete-button-disabled"
                  disabled
                >
                  削除
                </button>
              )}
            </div>
          ))}
        </div>
      )}
      <div>{isLimitReached && <p>スレッド数の上限に達しています</p>}</div>
      <div className="pagination-container">
        {items.map(({ type, page, selected, disabled, onClick }, index) => (
          <button
            key={index}
            onClick={onClick}
            selected={selected}
            disabled={disabled}
            type="button"
            className={`pagination-button ${selected ? "selected" : ""}`}
          >
            {getLabel(type, page)}
          </button>
        ))}
      </div>
    </div>
  );
}

export default Home;
