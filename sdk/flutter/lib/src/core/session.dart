class Session{String?_at,_rt,_uid;DateTime?_exp;
  String? get accessToken=>_at;String? get refreshToken=>_rt;String? get userId=>_uid;
  bool get isAuthenticated=>_at!=null&&!isExpired;
  bool get isExpired=>_exp!=null&&DateTime.now().isAfter(_exp!);
  void onLogin({required String accessToken,required String refreshToken,required DateTime expiresAt,required String userId})
    {_at=accessToken;_rt=refreshToken;_exp=expiresAt;_uid=userId;}
  void clear(){_at=null;_rt=null;_exp=null;_uid=null;}
}
