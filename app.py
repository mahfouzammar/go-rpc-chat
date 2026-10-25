import os

import requests
import streamlit as st

st.set_page_config(page_title="Go Chat", page_icon="💬")


def get_server_url() -> str:
    # على Streamlit Cloud: حطه في Settings -> Secrets  (SERVER_URL = "https://...")
    try:
        return st.secrets["SERVER_URL"].rstrip("/")
    except Exception:
        return os.environ.get("SERVER_URL", "http://localhost:1234").rstrip("/")


SERVER_URL = get_server_url()


def api(method: str, path: str, **kwargs):
    r = requests.request(method, SERVER_URL + path, timeout=8, **kwargs)
    r.raise_for_status()
    return r.json()


def reset_session(message: str | None = None):
    st.session_state.name = None
    st.session_state.messages = []
    if message:
        st.session_state.notice = message


# ---------- state ----------
st.session_state.setdefault("name", None)
st.session_state.setdefault("messages", [])
st.session_state.setdefault("notice", None)

# ---------- شاشة الدخول ----------
if not st.session_state.name:
    st.title("💬 Go Chat")

    if st.session_state.notice:
        st.warning(st.session_state.notice)
        st.session_state.notice = None

    with st.form("join_form"):
        typed_name = st.text_input("اكتب اسمك / Enter your name")
        submitted = st.form_submit_button("Join")

    if submitted:
        typed_name = typed_name.strip()
        if not typed_name:
            st.warning("لازم تكتب اسمك الأول.")
        else:
            try:
                res = api("POST", "/api/join", json={"name": typed_name})
            except requests.RequestException:
                st.error("مش قادر أوصل للسيرفر. اتأكد إنه شغال.")
            else:
                if res["status"] == "EXISTS":
                    st.error("الاسم ده مستخدم، اختار اسم تاني.")
                else:
                    st.session_state.name = typed_name
                    st.rerun()
    st.stop()

# ---------- شاشة الشات ----------
me = st.session_state.name
st.title("💬 Go Chat")

with st.sidebar:
    st.markdown(f"أنت داخل باسم: **{me}**")
    if st.button("Leave"):
        try:
            api("POST", "/api/leave", json={"name": me})
        except requests.RequestException:
            pass
        reset_session()
        st.rerun()


@st.fragment(run_every=2)
def chat_view():
    # بنسحب الرسائل الجديدة كل ثانيتين
    try:
        new_messages = api("GET", "/api/poll", params={"name": me})
    except requests.HTTPError as e:
        if e.response is not None and e.response.status_code == 404:
            reset_session("الجلسة خلصت، ادخل اسمك تاني.")
            st.rerun()
        new_messages = []
    except requests.RequestException:
        st.warning("فقدت الاتصال بالسيرفر... بحاول تاني.")
        new_messages = []

    for m in new_messages:
        if m["user"] != me:  # رسائلي أنا بتتضاف محلياً لحظة الإرسال
            st.session_state.messages.append(m)

    for m in st.session_state.messages:
        if m["user"] == "System":
            st.caption(f"🔔 {m['body']}")
        else:
            mine = m["user"] == me
            with st.chat_message(m["user"], avatar="🙂" if mine else "👤"):
                st.markdown(f"**{m['user']}**")
                st.write(m["body"])


prompt = st.chat_input("اكتب رسالتك...")
if prompt:
    try:
        api("POST", "/api/send", json={"user": me, "body": prompt})
    except requests.RequestException:
        st.error("الرسالة ماتبعتتش.")
    else:
        st.session_state.messages.append({"user": me, "body": prompt})

chat_view()