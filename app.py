import json
import urllib.request
import streamlit as st

GO_SERVER_URL = "http://localhost:1234/messages"

st.set_page_config(page_title="Go Chat Room", page_icon="💬", layout="centered")


def fetch_messages_from_go():
    try:
        req = urllib.request.Request(GO_SERVER_URL, method="GET")
        with urllib.request.urlopen(req, timeout=3) as response:
            return json.loads(response.read().decode("utf-8"))
    except Exception:
        return []


def send_message_to_go(sender, text):
    try:
        payload = json.dumps({"Sender": sender, "Text": text}).encode("utf-8")
        req = urllib.request.Request(
            GO_SERVER_URL,
            data=payload,
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with urllib.request.urlopen(req, timeout=3) as response:
            return response.status == 200
    except Exception:
        return False


if "username" not in st.session_state:
    st.session_state.username = None

# --- الشاشة الأولى: تسجيل الدخول ---
if not st.session_state.username:
    st.title("💬 غرفة المحادثة الجماعية")
    st.subheader("أدخل اسمك للانضمام إلى الغرفة")

    input_name = st.text_input("اسم المستخدم:", placeholder="اكتب اسمك هنا...")

    if st.button("انضمام 🚀"):
        if input_name.strip():
            st.session_state.username = input_name.strip()
            send_message_to_go(
                "System",
                f"*** {st.session_state.username} انضم إلى الغرفة ***",
            )
            st.rerun()
        else:
            st.warning("رجاءً اكتب اسم مستخدم صحيح.")

# --- الشاشة الثانية: واجهة الشات ---
else:
    st.title("💬 غرفة المحادثة الجماعية")
    st.caption(f"متصل باسم: **{st.session_state.username}**")

    if st.button("🔄 تحديث الرسائل"):
        st.rerun()

    st.divider()

    # جلب الرسائل وعرضها للجميع بالتساوي
    messages = fetch_messages_from_go()

    for msg in messages:
        sender = msg.get("Sender", "")
        text = msg.get("Text", "")

        if sender == "System":
            st.info(text)
        else:
            is_me = sender == st.session_state.username
            role_type = "user" if is_me else "assistant"

            with st.chat_message(role_type):
                st.write(f"**{sender}**: {text}")

    # إرسال رسالة جديدة
    if prompt := st.chat_input("اكتب رسالتك هنا..."):
        send_message_to_go(st.session_state.username, prompt)
        st.rerun()
