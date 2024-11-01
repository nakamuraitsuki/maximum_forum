export default function Register() {
  return (
    <div>
      <h1>Register</h1>
      <form>
        <div>
          <label htmlFor="name">Name</label>
          <input type="text" id="name" placeholder="Enter name" />

          <label htmlFor="password">Password</label>
          <input type="password" id="password" placeholder="Enter password" />

          <label htmlFor="confirmPassword">Confirm Password</label>
          <input type="password" id="confirmPassword" placeholder="Enter password again" />

          <button type="submit">Register</button>
          Already have an account? <a href="/login">Login</a>
        </div>
      </form>
    </div>
  );
}