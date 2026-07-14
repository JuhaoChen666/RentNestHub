import { MessageCircle, Send } from "lucide-react";
import { useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import type { InquiryMessage, User } from "@/types";

type Conversation = {
  key: string;
  houseId: number;
  houseTitle: string;
  otherUser: User;
  messages: InquiryMessage[];
};

export function MessageList({
  messages,
  currentUser,
  onSend,
}: {
  messages: InquiryMessage[];
  currentUser: User;
  onSend: (houseId: number, recipientId: number, content: string) => Promise<void>;
}) {
  const conversations = useMemo(() => {
    const grouped = new Map<string, Conversation>();
    for (const message of messages) {
      const otherUser = message.sender.id === currentUser.id ? message.recipient : message.sender;
      const key = `${message.houseId}:${otherUser.id}`;
      const conversation = grouped.get(key) ?? {
        key,
        houseId: message.houseId,
        houseTitle: message.houseTitle,
        otherUser,
        messages: [],
      };
      conversation.messages.push(message);
      grouped.set(key, conversation);
    }
    return [...grouped.values()].sort((left, right) => {
      const leftTime = left.messages.at(-1)?.createdAt ?? "";
      const rightTime = right.messages.at(-1)?.createdAt ?? "";
      return rightTime.localeCompare(leftTime);
    });
  }, [currentUser.id, messages]);
  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [content, setContent] = useState("");
  const [sending, setSending] = useState(false);
  const [error, setError] = useState("");
  const selected = conversations.find((conversation) => conversation.key === selectedKey);

  if (messages.length === 0) {
    return (
      <div className="empty-state">
        <MessageCircle size={28} />
        <h3>还没有咨询记录</h3>
        <p>向房东留言后，记录会显示在这里。</p>
      </div>
    );
  }

  return (
    <div className="message-list">
      {conversations.map((conversation) => (
        <Card className="message-record" key={conversation.key}>
          <CardContent>
            <button className="conversation-summary" onClick={() => setSelectedKey(conversation.key)} type="button">
              <div className="message-record-head">
                <strong>{conversation.houseTitle}</strong>
                <time>{new Date(conversation.messages.at(-1)?.createdAt ?? "").toLocaleDateString("zh-CN")}</time>
              </div>
              <span>{conversation.otherUser.displayName} · {conversation.otherUser.email}</span>
              <p>{conversation.messages.at(-1)?.content}</p>
            </button>
          </CardContent>
        </Card>
      ))}
      {selected && (
        <section className="conversation-detail">
          <div className="conversation-detail-head">
            <div><strong>{selected.houseTitle}</strong><span>{selected.otherUser.displayName} · {selected.otherUser.email}</span></div>
            <Button onClick={() => setSelectedKey(null)} size="sm" type="button" variant="ghost">返回</Button>
          </div>
          <div className="chat-history">
            {selected.messages.map((message) => (
              <div className={`chat-bubble ${message.sender.id === currentUser.id ? "is-own" : ""}`} key={message.id}>
                <span>{message.sender.displayName}</span>
                <p>{message.content}</p>
                <time>{new Date(message.createdAt).toLocaleString("zh-CN")}</time>
              </div>
            ))}
          </div>
          <form className="chat-composer" onSubmit={async (event) => {
            event.preventDefault();
            if (!content.trim()) return;
            setSending(true);
            setError("");
            try {
              await onSend(selected.houseId, selected.otherUser.id, content.trim());
              setContent("");
            } catch (sendError) {
              setError(sendError instanceof Error ? sendError.message : "发送失败");
            } finally {
              setSending(false);
            }
          }}>
            <Textarea maxLength={1000} onChange={(event) => setContent(event.target.value)} placeholder="输入回复内容" value={content} />
            <Button disabled={sending || !content.trim()} size="icon" type="submit"><Send size={16} /></Button>
          </form>
          {error && <p className="form-error">{error}</p>}
        </section>
      )}
    </div>
  );
}
