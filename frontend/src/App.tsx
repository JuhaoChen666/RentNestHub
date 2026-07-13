import {
  BedDouble,
  Bot,
  Building2,
  Check,
  ChevronDown,
  Heart,
  Home,
  MapPin,
  MessageCircle,
  Plus,
  Search,
  Send,
  SlidersHorizontal,
  Sparkles,
  Upload,
  X,
} from "lucide-react";
import {
  type FormEvent,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  favoriteHouse,
  listInquiryMessages,
  listFavoriteHouses,
  listHouses,
  publishHouse,
  recommend,
  sendMessage,
  unfavoriteHouse,
} from "./api";
import { MessageList } from "@/components/MessageList";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { NativeSelect } from "@/components/ui/native-select";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import type {
  House,
  HouseFilters,
  InquiryMessage,
  ListingMeta,
  Recommendation,
  RecommendationResult,
} from "./types";

const fallbackImage =
  "https://images.unsplash.com/photo-1560185008-b033106af5c3?auto=format&fit=crop&w=1200&q=80";

const initialFilters: HouseFilters = {
  city: "上海",
  district: "",
  keyword: "",
  maxRent: "",
  bedrooms: "",
};

const initialListingMeta: ListingMeta = {
  limit: 24,
  offset: 0,
  count: 0,
  hasMore: false,
  sort: "latest",
};

function App() {
  const [filters, setFilters] = useState(initialFilters);
  const [houses, setHouses] = useState<House[]>([]);
  const [listingMeta, setListingMeta] = useState(initialListingMeta);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [inquiryMessages, setInquiryMessages] = useState<InquiryMessage[]>([]);
  const [favoriteHouses, setFavoriteHouses] = useState<House[]>([]);
  const [recommendationMode, setRecommendationMode] = useState("");
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");
  const [activeView, setActiveView] = useState<"browse" | "recommend" | "favorites" | "messages">("browse");
  const [favorites, setFavorites] = useState<Set<number>>(new Set());
  const [publishOpen, setPublishOpen] = useState(false);
  const [messageHouse, setMessageHouse] = useState<House | null>(null);

  const loadHouses = useCallback(
    async (nextFilters: HouseFilters, offset = 0, append = false) => {
      if (append) {
        setLoadingMore(true);
      } else {
        setLoading(true);
      }
      setError("");
      try {
        const result = await listHouses(nextFilters, offset);
        setHouses((current) =>
          append ? [...current, ...result.items] : result.items,
        );
        setListingMeta(result.meta);
      } catch (loadError) {
        setError(
          loadError instanceof Error ? loadError.message : "房源加载失败",
        );
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [],
  );

  useEffect(() => {
    void loadHouses(initialFilters);
  }, [loadHouses]);

  const visibleHouses = useMemo(
    () =>
      activeView === "browse"
        ? houses
        : activeView === "favorites"
          ? favoriteHouses
          : recommendations.map((item) => item.house),
    [activeView, favoriteHouses, houses, recommendations],
  );

  async function handleSearch(event: FormEvent) {
    event.preventDefault();
    setActiveView("browse");
    setRecommendationMode("");
    await loadHouses(filters, 0);
  }

  async function showFavorites() {
    setActiveView("favorites");
    setLoading(true);
    setError("");
    try {
      const items = await listFavoriteHouses();
      setFavoriteHouses(items);
      setFavorites(new Set(items.map((house) => house.id)));
    } catch (favoritesError) {
      setError(
        favoritesError instanceof Error ? favoritesError.message : "收藏加载失败",
      );
    } finally {
      setLoading(false);
    }
  }

  async function handleLoadMore() {
    await loadHouses(filters, listingMeta.offset + listingMeta.count, true);
  }

  async function showMessages() {
    setActiveView("messages");
    setLoading(true);
    setError("");
    try {
      setInquiryMessages(await listInquiryMessages());
    } catch (messagesError) {
      setError(
        messagesError instanceof Error ? messagesError.message : "消息加载失败",
      );
    } finally {
      setLoading(false);
    }
  }

  async function toggleFavorite(houseId: number) {
    if (favorites.has(houseId)) {
      try {
        await unfavoriteHouse(houseId);
        setFavorites((current) => {
          const next = new Set(current);
          next.delete(houseId);
          return next;
        });
        setFavoriteHouses((current) => current.filter((house) => house.id !== houseId));
      } catch (favoriteError) {
        setError(
          favoriteError instanceof Error ? favoriteError.message : "取消收藏失败",
        );
      }
      return;
    }
    try {
      await favoriteHouse(houseId);
      setFavorites((current) => new Set(current).add(houseId));
    } catch (favoriteError) {
      setError(
        favoriteError instanceof Error
          ? favoriteError.message
          : "收藏失败",
      );
    }
  }

  return (
    <div className="app-shell">
      <Header
        onFavorites={() => void showFavorites()}
        onMessages={() => void showMessages()}
        onPublish={() => setPublishOpen(true)}
      />

      <main>
        <section className="workspace-head">
          <div>
            <p className="eyebrow">上海租房工作台</p>
            <h1>找到更适合你的下一处住所</h1>
            <p className="intro">
              按区域和预算快速筛选，也可以把真实需求交给智能匹配。
            </p>
          </div>
          <div className="market-note">
            <Building2 size={18} />
            <span>当前展示真实可租房源</span>
          </div>
        </section>

        <form className="filter-bar" onSubmit={handleSearch}>
          <label className="search-field">
            <Search size={18} />
            <Input
              aria-label="关键词"
              value={filters.keyword}
              onChange={(event) =>
                setFilters({ ...filters, keyword: event.target.value })
              }
              placeholder="地铁、通勤、采光或小区"
            />
          </label>
          <SelectField
            icon={<MapPin size={17} />}
            label="区域"
            value={filters.district}
            onChange={(value) => setFilters({ ...filters, district: value })}
            options={["", "徐汇区", "浦东新区", "静安区", "闵行区"]}
          />
          <SelectField
            icon={<span className="yuan">¥</span>}
            label="月租上限"
            value={filters.maxRent}
            onChange={(value) => setFilters({ ...filters, maxRent: value })}
            options={["", "4000", "5000", "6000", "8000", "10000"]}
            format={(value) => (value ? `¥${value}` : "不限")}
          />
          <SelectField
            icon={<BedDouble size={17} />}
            label="卧室"
            value={filters.bedrooms}
            onChange={(value) => setFilters({ ...filters, bedrooms: value })}
            options={["", "1", "2", "3"]}
            format={(value) => (value ? `${value} 室及以上` : "不限")}
          />
          <Button className="primary-button search-button" type="submit">
            <SlidersHorizontal size={18} />
            筛选房源
          </Button>
        </form>

        <div className="content-layout">
          <aside className="recommend-panel">
            <div className="panel-icon">
              <Bot size={22} />
            </div>
            <p className="eyebrow">AI 智能匹配</p>
            <h2>说说你理想的房子</h2>
            <p className="panel-copy">
              可以写通勤、宠物、采光、做饭等偏好，我们会结合预算与房源条件排序。
            </p>
            <RecommendationForm
              filters={filters}
              onComplete={(result) => {
                setRecommendations(result.items);
                setRecommendationMode(result.mode);
                setActiveView("recommend");
              }}
              onError={setError}
            />
          </aside>

          <section className="results" aria-live="polite">
            <div className="results-head">
              <div>
                <p className="eyebrow">
                  {activeView === "browse"
                    ? "筛选结果"
                    : activeView === "favorites"
                      ? "我的收藏"
                      : activeView === "messages"
                        ? "我的消息"
                      : "专属推荐"}
                </p>
                <h2>
                  {activeView === "browse"
                    ? `${visibleHouses.length} 套可选房源`
                    : activeView === "favorites"
                      ? `${visibleHouses.length} 套已收藏房源`
                      : activeView === "messages"
                        ? `${inquiryMessages.length} 条咨询记录`
                      : "根据你的需求排序"}
                </h2>
                {activeView === "recommend" && recommendationMode && (
                  <Badge className="recommend-mode-badge">
                    {recommendationModeLabel(recommendationMode)}
                  </Badge>
                )}
              </div>
              {activeView !== "browse" && (
                <Button
                  className="text-button"
                  onClick={() => setActiveView("browse")}
                  type="button"
                  variant="ghost"
                >
                  返回筛选结果
                </Button>
              )}
            </div>

            {error && (
              <div className="error-banner">
                <span>{error}</span>
                <Button
                  aria-label="关闭错误提示"
                  onClick={() => setError("")}
                  type="button"
                  variant="ghost"
                  size="icon"
                >
                  <X size={16} />
                </Button>
              </div>
            )}

            {loading ? (
              <div className="loading-grid">
                {[1, 2, 3].map((item) => (
                  <Skeleton className="skeleton" key={item} />
                ))}
              </div>
            ) : activeView === "messages" ? (
              <MessageList messages={inquiryMessages} />
            ) : visibleHouses.length > 0 ? (
              <>
                <div className="house-grid">
                  {visibleHouses.map((house) => (
                    <HouseCard
                      favorite={favorites.has(house.id)}
                      house={house}
                      key={house.id}
                      recommendation={
                        activeView === "recommend"
                          ? recommendations.find((item) => item.house.id === house.id)
                          : undefined
                      }
                      onFavorite={() => void toggleFavorite(house.id)}
                      onMessage={() => setMessageHouse(house)}
                    />
                  ))}
                </div>
                {activeView === "browse" && listingMeta.hasMore && (
                  <div className="load-more-row">
                    <Button
                      className="load-more-button"
                      disabled={loadingMore}
                      onClick={() => void handleLoadMore()}
                      type="button"
                      variant="outline"
                    >
                      {loadingMore ? "正在加载..." : "加载更多"}
                    </Button>
                  </div>
                )}
              </>
            ) : (
              <div className="empty-state">
                <Home size={28} />
                <h3>没有找到匹配房源</h3>
                <p>放宽区域或预算条件后再试一次。</p>
              </div>
            )}
          </section>
        </div>
      </main>

      {publishOpen && (
        <PublishDialog
          onClose={() => setPublishOpen(false)}
          onPublished={(house) => {
            setHouses((current) => [house, ...current]);
            setPublishOpen(false);
          }}
        />
      )}
      {messageHouse && (
        <MessageDialog
          house={messageHouse}
          onClose={() => setMessageHouse(null)}
        />
      )}
    </div>
  );
}

function Header({
  onFavorites,
  onMessages,
  onPublish,
}: {
  onFavorites: () => void;
  onMessages: () => void;
  onPublish: () => void;
}) {
  return (
    <header className="topbar">
      <a className="brand" href="/" aria-label="RentNestHub 首页">
        <span className="brand-mark">
          <Home size={19} />
        </span>
        <span>RentNestHub</span>
      </a>
      <nav aria-label="主导航">
        <a className="active" href="#homes">
          找房
        </a>
        <a href="#favorites" onClick={(event) => { event.preventDefault(); onFavorites(); }}>收藏</a>
        <a href="#messages" onClick={(event) => { event.preventDefault(); onMessages(); }}>消息</a>
      </nav>
      <div className="topbar-actions">
        <Button
          className="secondary-button"
          onClick={onPublish}
          type="button"
          variant="secondary"
        >
          <Plus size={17} />
          发布房源
        </Button>
        <Button className="avatar" aria-label="用户中心" type="button" size="icon">
          陈
        </Button>
      </div>
    </header>
  );
}

function SelectField({
  icon,
  label,
  value,
  options,
  onChange,
  format = (option) => option || "不限",
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  options: string[];
  onChange: (value: string) => void;
  format?: (value: string) => string;
}) {
  return (
    <label className="select-field">
      {icon}
      <span className="sr-only">{label}</span>
      <NativeSelect
        aria-label={label}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {options.map((option) => (
          <option key={option || "all"} value={option}>
            {format(option)}
          </option>
        ))}
      </NativeSelect>
      <ChevronDown size={15} />
    </label>
  );
}

function RecommendationForm({
  filters,
  onComplete,
  onError,
}: {
  filters: HouseFilters;
  onComplete: (result: RecommendationResult) => void;
  onError: (message: string) => void;
}) {
  const [need, setNeed] = useState(
    "离地铁近，采光好，可以做饭，最好有独立阳台",
  );
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setSubmitting(true);
    onError("");
    try {
      onComplete(
        await recommend({
          need,
          city: filters.city,
          district: filters.district,
          maxRent: Number(filters.maxRent) || 7000,
          bedrooms: Number(filters.bedrooms) || 1,
        }),
      );
    } catch (recommendError) {
      onError(
        recommendError instanceof Error
          ? recommendError.message
          : "推荐生成失败",
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <Textarea
        aria-label="租房需求"
        maxLength={500}
        value={need}
        onChange={(event) => setNeed(event.target.value)}
      />
      <Button
        className="ai-button"
        disabled={submitting || !need.trim()}
        type="submit"
      >
        <Sparkles size={18} />
        {submitting ? "正在匹配..." : "生成专属推荐"}
      </Button>
    </form>
  );
}

function recommendationModeLabel(mode: string) {
  return mode === "ai-http" ? "AI 服务推荐" : "本地规则推荐";
}

function HouseCard({
  house,
  recommendation,
  favorite,
  onFavorite,
  onMessage,
}: {
  house: House;
  recommendation?: Recommendation;
  favorite: boolean;
  onFavorite: () => void;
  onMessage: () => void;
}) {
  return (
    <Card className="house-card">
      <div className="house-image-wrap">
        <img
          className="house-image"
          src={house.imageUrls[0] || fallbackImage}
          alt={`${house.title}室内实景`}
        />
        <Button
          className={`favorite-button ${favorite ? "is-favorite" : ""}`}
          aria-label={favorite ? "取消收藏" : "收藏房源"}
          onClick={onFavorite}
          type="button"
          size="icon"
          variant="ghost"
        >
          <Heart size={18} fill={favorite ? "currentColor" : "none"} />
        </Button>
        {recommendation && (
          <Badge className="match-badge">
            {Math.round(recommendation.score)}% 匹配
          </Badge>
        )}
      </div>
      <CardContent className="house-body">
        <div className="house-location">
          <MapPin size={14} />
          {house.city} · {house.district}
        </div>
        <h3>{house.title}</h3>
        <div className="house-facts">
          <span>{house.bedrooms} 室 {house.bathrooms} 卫</span>
          <span>{house.areaSqm} m²</span>
        </div>
        {recommendation && (
          <p className="recommend-reason">
            <Sparkles size={14} />
            {recommendation.reason}
          </p>
        )}
        <div className="amenity-list">
          {house.amenities.slice(0, 3).map((amenity) => (
            <Badge key={amenity}>{amenity}</Badge>
          ))}
        </div>
        <div className="house-footer">
          <div className="price">
            <strong>¥{house.monthlyRent.toLocaleString()}</strong>
            <span>/ 月</span>
          </div>
          <Button
            className="icon-text-button"
            onClick={onMessage}
            type="button"
            variant="outline"
          >
            <MessageCircle size={16} />
            咨询
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function DialogFrame({
  title,
  children,
  onClose,
}: {
  title: string;
  children: React.ReactNode;
  onClose: () => void;
}) {
  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="dialog">
        <DialogHeader className="dialog-head">
          <DialogTitle id="dialog-title">{title}</DialogTitle>
        </DialogHeader>
        {children}
      </DialogContent>
    </Dialog>
  );
}

function PublishDialog({
  onClose,
  onPublished,
}: {
  onClose: () => void;
  onPublished: (house: House) => void;
}) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError("");
    const form = new FormData(event.currentTarget);
    form.set("landlordId", "1");
    try {
      onPublished(await publishHouse(form));
    } catch (publishError) {
      setError(
        publishError instanceof Error ? publishError.message : "发布失败",
      );
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <DialogFrame title="发布新房源" onClose={onClose}>
      <form className="dialog-form" onSubmit={handleSubmit}>
        <label className="wide">
          房源标题
          <Input
            maxLength={120}
            name="title"
            required
            placeholder="例如：徐汇滨江明亮两居"
          />
        </label>
        <label>
          城市
          <Input maxLength={80} name="city" required defaultValue="上海" />
        </label>
        <label>
          区域
          <Input maxLength={80} name="district" required placeholder="徐汇区" />
        </label>
        <label className="wide">
          详细地址
          <Input
            maxLength={180}
            name="address"
            required
            placeholder="仅用于房源信息展示"
          />
        </label>
        <label>
          月租（元）
          <Input
            inputMode="numeric"
            max="200000"
            min="1"
            name="monthlyRent"
            required
            type="number"
          />
        </label>
        <label>
          面积（m²）
          <Input
            inputMode="decimal"
            max="2000"
            min="1"
            name="areaSqm"
            required
            step="0.1"
            type="number"
          />
        </label>
        <label>
          卧室
          <Input
            inputMode="numeric"
            max="20"
            min="1"
            name="bedrooms"
            required
            type="number"
          />
        </label>
        <label>
          卫生间
          <Input
            defaultValue="1"
            inputMode="numeric"
            max="20"
            min="1"
            name="bathrooms"
            required
            type="number"
          />
        </label>
        <label className="wide">
          配套设施
          <Input
            maxLength={360}
            name="amenities"
            placeholder="近地铁, 电梯, 可做饭"
          />
        </label>
        <label className="wide">
          房源描述
          <Textarea
            maxLength={1000}
            name="description"
            required
            placeholder="介绍采光、交通和入住条件"
          />
        </label>
        <label className="upload-field wide">
          <Upload size={18} />
          上传房源图片
          <Input
            accept="image/jpeg,image/png,image/webp"
            aria-label="上传房源图片，最多 8 张，每张不超过 5 MB"
            multiple
            name="images"
            type="file"
          />
        </label>
        {error && <p className="form-error">{error}</p>}
        <div className="dialog-actions wide">
          <Button
            className="text-button"
            onClick={onClose}
            type="button"
            variant="ghost"
          >
            取消
          </Button>
          <Button className="primary-button" disabled={submitting} type="submit">
            <Plus size={17} />
            {submitting ? "正在发布..." : "发布房源"}
          </Button>
        </div>
      </form>
    </DialogFrame>
  );
}

function MessageDialog({
  house,
  onClose,
}: {
  house: House;
  onClose: () => void;
}) {
  const [content, setContent] = useState(
    `你好，我对「${house.title}」感兴趣，请问近期方便看房吗？`,
  );
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    try {
      await sendMessage(house.id, content);
      setSent(true);
    } catch (messageError) {
      setError(
        messageError instanceof Error ? messageError.message : "留言发送失败",
      );
    }
  }

  return (
    <DialogFrame title="咨询房东" onClose={onClose}>
      {sent ? (
        <div className="success-state">
          <span><Check size={24} /></span>
          <h3>留言已发送</h3>
          <p>房东回复后会出现在站内消息中。</p>
          <Button className="primary-button" onClick={onClose} type="button">
            完成
          </Button>
        </div>
      ) : (
        <form className="message-form" onSubmit={handleSubmit}>
          <div className="message-house">
            <img src={house.imageUrls[0] || fallbackImage} alt="" />
            <div>
              <strong>{house.title}</strong>
              <span>¥{house.monthlyRent.toLocaleString()} / 月</span>
            </div>
          </div>
          <label>
            留言内容
            <Textarea
              maxLength={1000}
              required
              value={content}
              onChange={(event) => setContent(event.target.value)}
            />
          </label>
          {error && <p className="form-error">{error}</p>}
          <Button className="primary-button" type="submit">
            <Send size={17} />
            发送留言
          </Button>
        </form>
      )}
    </DialogFrame>
  );
}

export default App;
