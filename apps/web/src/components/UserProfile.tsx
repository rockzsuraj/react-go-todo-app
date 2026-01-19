type Props = {
  user: {
    name: string;
    email: string;
    picture?: string;
  };
};

export default function UserProfile({ user }: Props) {
  return (
    <div className="d-flex align-items-center gap-3">
      <img
        src={user.picture || `https://ui-avatars.com/api/?name=${user.name}`}
        alt={user.name}
        width={36}
        height={36}
        className="rounded-circle border"
      />

      <div>
        <div className="fw-semibold">{user.name}</div>
        <small className="text-muted">{user.email}</small>
      </div>
    </div>
  );
}