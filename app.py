import streamlit as st

st.set_page_config(page_title="Go RPC Chat", page_icon="💬")

st.title("💬 غرفة المحادثة - Chat Room")

# 1. تهيئة سجل الرسائل في الجلسة إذا لم يكن موجوداً
if "messages" not in st.session_state:
    st.session_state.messages = []

# 2. إدخال اسم المستخدم
username = st.text_input("اسم المستخدم", value="User", key="username_input")

st.divider()

# 3. عرض جميع الرسائل السابقة المطبوعة في السجل
for msg in st.session_state.messages:
    with st.chat_message(msg["role"]):
        st.write(f"**{msg['sender']}**: {msg['text']}")

# 4. مربع إرسال الرسائل الجديد (مثل تطبيقات الشات)
if prompt := st.chat_input("اكتب رسالتك هنا..."):
    if username.strip() == "":
        st.warning("يرجى إدخال اسم المستخدم أولاً.")
    else:
        # إضافة الرسالة إلى السجل المحلي
        new_msg = {"role": "user", "sender": username, "text": prompt}
        st.session_state.messages.append(new_msg)
        
        # إعادة تحديث الواجهة فوراً لعرض الرسالة الجديدة
        st.rerun()