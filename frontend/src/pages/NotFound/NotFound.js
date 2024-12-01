import { Link } from "react-router-dom";
import "./NotFound.css";

export default function NotFound() {
  return (
    <div className="not-found">
      <p>このページのアクセス権限がありません</p>
      <Link to="/">スレッド一覧</Link>
    </div>
  );
}
