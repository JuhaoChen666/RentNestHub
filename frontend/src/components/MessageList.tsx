import { MessageCircle } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";
import type { InquiryMessage } from "@/types";

export function MessageList({ messages }: { messages: InquiryMessage[] }) {
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
      {messages.map((message) => (
        <Card className="message-record" key={message.id}>
          <CardContent>
            <div className="message-record-head">
              <strong>{message.houseTitle}</strong>
              <time>{new Date(message.createdAt).toLocaleDateString("zh-CN")}</time>
            </div>
            <p>{message.content}</p>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
