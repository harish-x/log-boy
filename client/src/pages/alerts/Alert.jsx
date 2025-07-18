import React, { useState, useEffect, useRef } from "react";
import {
  useLazyGetAllAlertRulesQuery,
  useLazyGetVerifiedEmailsQuery,
  useCreateAlertEmailMutation,
  useVerifyAlertEmailMutation,
  useCreateAlertRuleMutation,
} from "@/services/AlertServices";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Checkbox } from "@/components/ui/checkbox";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";
import {
  Bell,
  Mail,
  Webhook,
  Plus,
  Eye,
  AlertTriangle,
  Info,
  ShieldAlert,
  CpuIcon,
  MemoryStickIcon,
  ServerCrashIcon,
  OctagonXIcon,
  SlackIcon,
} from "lucide-react";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { useParams } from "react-router-dom";

const AlertManagement = () => {
  const { projectName } = useParams();
  const [alertRules, setAlertRules] = useState([]);
  const [verifiedEmails, setVerifiedEmails] = useState([]);
  const [selectedRuleType, setSelectedRuleType] = useState("");
  const [alertForm, setAlertForm] = useState({
    metric_name: "",
    log_field: "",
    event_type: "",
    operator: ">=",
    threshold: "",
    time_window: "5 minutes",
    severity: "warning",
    project_name: "project_1",
    alert_methods: [],
  });
  const [emailForm, setEmailForm] = useState({
    email: "",
    project_name: projectName,
  });
  const [otpForm, setOtpForm] = useState({
    email: "",
    project: projectName,
    otp: "",
  });
  const [selectedEmails, setSelectedEmails] = useState("");
  const [webhookUrl, setWebhookUrl] = useState("");
  const [showEmailDialog, setShowEmailDialog] = useState(false);
  const [showOtpDialog, setShowOtpDialog] = useState(false);

  // RTK Query hooks
  const [getAllAlertRules, { data: alertRulesData, isLoading: alertRulesLoading }] = useLazyGetAllAlertRulesQuery();
  const [getVerifiedEmails, { data: verifiedEmailsData, isLoading: emailsLoading }] = useLazyGetVerifiedEmailsQuery();
  const [createAlertEmail, { isLoading: creatingEmail }] = useCreateAlertEmailMutation();
  const [verifyAlertEmail, { isLoading: verifyingEmail }] = useVerifyAlertEmailMutation();
  const [createAlertRule, { isLoading: creatingAlert }] = useCreateAlertRuleMutation();
  const [selectedMethods, setSelectedMethods] = useState({ webhook: false, email: false });
  const targetRef = useRef(null);

  useEffect(() => {
    fetchAlertRules();
    fetchVerifiedEmails();
  }, []);

  useEffect(() => {
    if (alertRulesData?.data) {
      setAlertRules(alertRulesData.data);
    }
  }, [alertRulesData]);

  useEffect(() => {
    if (verifiedEmailsData?.data) {
      setVerifiedEmails(verifiedEmailsData.data);
    }
  }, [verifiedEmailsData]);

  const fetchAlertRules = async () => {
    try {
      await getAllAlertRules(projectName);
    } catch (error) {
      console.error("Error fetching alert rules:", error);
    }
  };

  const fetchVerifiedEmails = async () => {
    try {
      await getVerifiedEmails(projectName);
    } catch (error) {
      console.error("Error fetching verified emails:", error);
    }
  };

  const handleCreateEmail = async () => {
    try {
      await createAlertEmail(emailForm).unwrap();
      setShowEmailDialog(false);
      setShowOtpDialog(true);
      setOtpForm({
        email: emailForm.email,
        project: emailForm.project_name,
        otp: "",
      });
    } catch (error) {
      console.error("Error creating email:", error);
    }
  };

  const handleVerifyEmail = async () => {
    try {
      await verifyAlertEmail(otpForm)
        .unwrap()
        .then((res) => setShowOtpDialog(false))
        .catch((error) => {
          toast.error(error?.data?.message, {
            variant: "destructive",
            richColors: true,
            action: <AlertTriangle className="text-red-500 ml-auto" />,
          });
          otpForm.otp = "";
        });

      fetchVerifiedEmails();
    } catch (error) {
      console.error("Error verifying email:", error);
    }
  };

  const handleCreateAlert = async () => {
    console.log("Creating alert:", alertForm);
    try {
      const alertMethods = [];

      // Add selected emails
      if (selectedMethods.email) {
        alertMethods.push({
          method: "email",
          value: selectedEmails,
        });
      }

      // Add webhook if provided
      if (webhookUrl) {
        alertMethods.push({
          method: "webhook",
          value: webhookUrl,
        });
      }

      const payload = {
        ...alertForm,
        threshold: parseFloat(alertForm.threshold),
        alert_methods: alertMethods,
      };

      // Set rule_type based on selected type
      if (selectedRuleType === "cpu_usage" || selectedRuleType === "memory_usage") {
        payload.rule_type = "metric_avg";
        payload.metric_name = selectedRuleType;
      } else if (selectedRuleType === "response_status") {
        payload.rule_type = "log_count";
        payload.log_field = "responseStatus";
      } else if (selectedRuleType === "server_restart") {
        payload.rule_type = "event_count";
        payload.event_type = "server_restart";
      }

      if (alertMethods.length === 0) {
        toast.error("Please select at least one alert method", {
          variant: "destructive",
          richColors: true,
          action: <AlertTriangle className="text-red-500 ml-auto" />,
        });
        return;
      }

      await createAlertRule(payload);

      // Reset form
      setAlertForm({
        metric_name: "",
        log_field: "",
        event_type: "",
        operator: ">=",
        threshold: "",
        time_window: "5 minutes",
        severity: "warning",
        project_name: "project_1",
        alert_methods: [],
      });
      setSelectedRuleType("");
      setSelectedEmails("");
      setWebhookUrl("");

      fetchAlertRules();
    } catch (error) {
      console.error("Error creating alert rule:", error);
    }
  };

  const theme = localStorage.getItem("vite-ui-theme");
  const getRuleTypeLabel = (ruleType) => {
    switch (ruleType) {
      case "cpu_usage":
        return "CPU Usage Monitoring";
      case "memory_usage":
        return "Memory Usage Monitoring";
      case "response_status":
        return "Response Status Monitoring";
      case "server_restart":
        return "Server Restart Detection";
      default:
        return "Unknown";
    }
  };

  const getSeverityColor = (severity) => {
    switch (severity) {
      case "critical":
        return "bg-red-500";
      case "warning":
        return "bg-yellow-500";
      case "info":
        return "bg-blue-500";
      default:
        return "bg-gray-500";
    }
  };

  const scrollToTarget = () => {
    if (targetRef.current) {
      targetRef.current.scrollIntoView({
        behavior: "smooth",
        block: "start",
      });
    }
  };

  return (
    <div className=" p-6 space-y-6 border border-primary/[0.20] px-2 w-[98%] mx-auto rounded-2xl">
      <div className="relative mb-8">
        <div className="relative flex items-center justify-center space-x-4 py-8">
          <div className="relative">
            <Bell className="relative h-12 w-12 text-white" />
          </div>
          <div className="text-center">
            <h1 className="text-4xl font-bold text-primary">Alert Management</h1>
            <p className="text-lg text-muted-foreground mt-2">Intelligent monitoring for your infrastructure</p>
          </div>
        </div>
      </div>

      <div className="relative">
        <div className="absolute inset-0 bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/10 dark:to-purple-900/10 blur-2xl"></div>
      </div>
      <Tabs defaultValue="create" className="relative w-full">
        <TabsList className="grid w-full grid-cols-2 backdrop-blur-sm border shadow-lg">
          <TabsTrigger value="create" className="data-[state=active]:bg-primary data-[state=active]:text-white">
            Create Alert
          </TabsTrigger>
          <TabsTrigger value="manage" className="data-[state=active]:bg-primary data-[state=active]:text-white">
            Manage Alerts
          </TabsTrigger>
        </TabsList>

        <TabsContent value="create" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Create New Alert Rule</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Rule Type Selection */}
              <div className="space-y-6">
                <div className="text-center space-y-2">
                  <Label className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                    Choose Your Alert Strategy
                  </Label>
                  <p className="text-muted-foreground">Select the monitoring rule that best fits your infrastructure needs</p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div
                    className={`group cursor-pointer transition-all duration-300 transform hover:scale-105 ${
                      selectedRuleType === "cpu_usage"
                        ? "ring-2 ring-blue-400 shadow-lg shadow-red-500/25"
                        : "hover:shadow-xl hover:shadow-red-500/10"
                    }`}
                    onClick={() => {
                      scrollToTarget();
                      setSelectedRuleType("cpu_usage");
                    }}
                  >
                    <Card className="h-full border-0 bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-800/20">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between mb-4">
                          <div className="flex items-center space-x-3">
                            <div className="relative">
                              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center shadow-lg">
                                <CpuIcon className=" text-blue-200" />
                              </div>
                              <div className="absolute -top-1 -right-1 w-4 h-4 bg-blue-400 rounded-full animate-bounce"></div>
                            </div>
                            <div>
                              <h3 className="font-bold text-lg text-blue-800 dark:text-blue-300">CPU Usage</h3>
                              <p className="text-sm text-blue-600 dark:text-blue-400">Performance Monitor</p>
                            </div>
                          </div>
                          {selectedRuleType === "cpu_usage" && (
                            <div className="w-6 h-6 rounded-full bg-blue-500 flex items-center justify-center">
                              <div className="w-3 h-3 rounded-full bg-white"></div>
                            </div>
                          )}
                        </div>
                        <p className="text-sm text-blue-700 dark:text-blue-300 mb-3">
                          Monitor and alert on CPU usage spikes across your infrastructure
                        </p>
                        <div className="flex items-center space-x-2 text-xs text-blue-600 dark:text-blue-400">
                          <span className="px-2 py-1 bg-blue-100 dark:bg-blue-800 rounded-full">Real-time</span>
                          <span className="px-2 py-1 bg-blue-100 dark:bg-blue-800 rounded-full">Threshold-based</span>
                        </div>
                      </CardContent>
                    </Card>
                  </div>

                  <div
                    className={`group cursor-pointer transition-all duration-300 transform hover:scale-105 ${
                      selectedRuleType === "memory_usage"
                        ? "ring-2 ring-green-500 shadow-lg shadow-green-500/25"
                        : "hover:shadow-xl hover:shadow-green-500/10"
                    }`}
                    onClick={() => {
                      setSelectedRuleType("memory_usage");
                      scrollToTarget();
                    }}
                  >
                    <Card className="h-full border-0 bg-gradient-to-br from-green-50 to-green-100 dark:from-green-900/20 dark:to-green-800/20">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between mb-4">
                          <div className="flex items-center space-x-3">
                            <div className="relative">
                              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-green-500 to-green-600 flex items-center justify-center shadow-lg">
                                <MemoryStickIcon className=" text-green-200" />
                              </div>
                              <div className="absolute -top-1 -right-1 w-4 h-4 bg-green-400 rounded-full animate-bounce"></div>
                            </div>
                            <div>
                              <h3 className="font-bold text-lg text-green-800 dark:text-green-300">Memory Usage</h3>
                              <p className="text-sm text-green-600 dark:text-green-400">Resource Monitor</p>
                            </div>
                          </div>
                          {selectedRuleType === "memory_usage" && (
                            <div className="w-6 h-6 rounded-full bg-green-500 flex items-center justify-center">
                              <div className="w-3 h-3 rounded-full bg-white"></div>
                            </div>
                          )}
                        </div>
                        <p className="text-sm text-green-700 dark:text-green-300 mb-3">Track memory consumption and prevent out-of-memory issues</p>
                        <div className="flex items-center space-x-2 text-xs text-green-600 dark:text-green-400">
                          <span className="px-2 py-1 bg-green-200 dark:bg-green-800 rounded-full">Predictive</span>
                          <span className="px-2 py-1 bg-green-200 dark:bg-green-800 rounded-full">Proactive</span>
                        </div>
                      </CardContent>
                    </Card>
                  </div>

                  <div
                    className={`group cursor-pointer transition-all duration-300 transform hover:scale-105 ${
                      selectedRuleType === "response_status"
                        ? "ring-2 ring-amber-500 shadow-lg shadow-amber-500/25"
                        : "hover:shadow-xl hover:shadow-amber-500/10"
                    }`}
                    onClick={() => {
                      setSelectedRuleType("response_status");
                      scrollToTarget();
                    }}
                  >
                    <Card className="h-full border-0 bg-gradient-to-br from-red-50 to-red-100 dark:from-red-900/20 dark:to-destructive/20">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between mb-4">
                          <div className="flex items-center space-x-3">
                            <div className="relative">
                              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-red-500 to-destructive flex items-center justify-center shadow-lg">
                                <OctagonXIcon className=" text-red-100" />
                              </div>
                              <div className="absolute -top-1 -right-1 w-4 h-4 bg-red-400 rounded-full animate-bounce"></div>
                            </div>
                            <div>
                              <h3 className="font-bold text-lg text-red-800 dark:text-red-300">Response Status</h3>
                              <p className="text-sm text-red-600 dark:text-red-400">Error Tracking</p>
                            </div>
                          </div>
                          {selectedRuleType === "response_status" && (
                            <div className="w-6 h-6 rounded-full bg-red-500 flex items-center justify-center">
                              <div className="w-3 h-3 rounded-full bg-white"></div>
                            </div>
                          )}
                        </div>
                        <p className="text-sm text-red-700 dark:text-red-300 mb-3">Monitor HTTP response codes and catch errors before users do</p>
                        <div className="flex items-center space-x-2 text-xs text-red-600 dark:text-red-400">
                          <span className="px-2 py-1 bg-red-100 dark:bg-red-800 rounded-full">Error Detection</span>
                          <span className="px-2 py-1 bg-red-100 dark:bg-red-800 rounded-full">User Experience</span>
                        </div>
                      </CardContent>
                    </Card>
                  </div>

                  <div
                    className={`group cursor-pointer transition-all duration-300 transform hover:scale-105 ${
                      selectedRuleType === "server_restart"
                        ? "ring-2 ring-purple-500 shadow-lg shadow-purple-500/25"
                        : "hover:shadow-xl hover:shadow-purple-500/10"
                    }`}
                    onClick={() => {
                      setSelectedRuleType("server_restart");
                      scrollToTarget();
                    }}
                  >
                    <Card className="h-full border-0 bg-gradient-to-br from-purple-50 to-purple-100 dark:from-purple-900/20 dark:to-purple-800/20">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between mb-4">
                          <div className="flex items-center space-x-3">
                            <div className="relative">
                              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500 to-purple-600 flex items-center justify-center shadow-lg">
                                <ServerCrashIcon className=" text-purple-100" />
                              </div>
                              <div className="absolute -top-1 -right-1 w-4 h-4 bg-purple-400 rounded-full animate-bounce"></div>
                            </div>
                            <div>
                              <h3 className="font-bold text-lg text-purple-800 dark:text-purple-300">Crash Detection</h3>
                              <p className="text-sm text-purple-600 dark:text-purple-400">Stability Monitor</p>
                            </div>
                          </div>
                          {selectedRuleType === "server_restart" && (
                            <div className="w-6 h-6 rounded-full bg-purple-500 flex items-center justify-center">
                              <div className="w-3 h-3 rounded-full bg-white"></div>
                            </div>
                          )}
                        </div>
                        <p className="text-sm text-purple-700 dark:text-purple-300 mb-3">
                          Detect crash loops and instability patterns in your services
                        </p>
                        <div className="flex items-center space-x-2 text-xs text-purple-600 dark:text-purple-400">
                          <span className="px-2 py-1 bg-purple-200 dark:bg-purple-800 rounded-full">Crash Detection</span>
                          <span className="px-2 py-1 bg-purple-200 dark:bg-purple-800 rounded-full">Stability</span>
                        </div>
                      </CardContent>
                    </Card>
                  </div>

                  <div className="relative group">
                    <Card className="h-full border-0 bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900/20 dark:to-gray-800/20 opacity-60">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between mb-4">
                          <div className="flex items-center space-x-3">
                            <div className="relative">
                              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-gray-400 to-gray-500 flex items-center justify-center shadow-lg">
                                <div className="w-6 h-6 rounded-full bg-white/20"></div>
                              </div>
                              <div className="absolute -top-1 -right-1 w-4 h-4 bg-gray-400 rounded-full"></div>
                            </div>
                            <div>
                              <h3 className="font-bold text-lg text-gray-600 dark:text-gray-400">Fraud Detection</h3>
                              <p className="text-sm text-gray-500 dark:text-gray-500">AI-Powered Security</p>
                            </div>
                          </div>
                        </div>
                        <p className="text-sm text-gray-600 dark:text-gray-400 mb-3">Advanced machine learning algorithms for anomaly detection</p>
                        <div className="flex items-center space-x-2 text-xs text-gray-500 dark:text-gray-500">
                          <span className="px-2 py-1 bg-gray-200 dark:bg-gray-700 rounded-full">AI/ML</span>
                          <span className="px-2 py-1 bg-gray-200 dark:bg-gray-700 rounded-full">Security</span>
                        </div>
                      </CardContent>
                    </Card>
                    <div className="absolute inset-0 flex items-center justify-center">
                      <div className="bg-black/80 text-white px-4 py-2 rounded-full text-sm font-medium">Coming Soon</div>
                    </div>
                  </div>
                </div>
              </div>

              {selectedRuleType && (
                <>
                  <Separator />
                  {/* Alert Configuration */}
                  <div className="space-y-6" ref={targetRef}>
                    <div className="text-center">
                      <Label className="text-2xl font-bold text-primary">Fine-tune Your Alert Parameters</Label>
                      <p className="text-muted-foreground mt-3">Configure thresholds and conditions with precision</p>
                    </div>

                    <div className="space-y-8 rounded-lg border bg-card text-card-foreground shadow-sm p-6">
                      <div className="space-y-4">
                        <h3 className="text-lg font-semibold">Define Alert Condition</h3>
                        <div className="p-4 border rounded-md flex flex-wrap items-center gap-3 bg-muted/40">
                          <span className="font-medium">When value is</span>

                          {/* Operator Dropdown */}
                          <Select value={alertForm.operator} onValueChange={(value) => setAlertForm({ ...alertForm, operator: value })}>
                            <SelectTrigger className="w-[180px]">
                              <SelectValue placeholder="Select operator" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value=">">Greater than</SelectItem>
                              <SelectItem value="<">Less than</SelectItem>
                              <SelectItem value=">=">Greater than or equal to</SelectItem>
                              <SelectItem value="<=">Less than or equal to</SelectItem>
                              <SelectItem value="==">Equal to</SelectItem>
                            </SelectContent>
                          </Select>

                          <div className="relative">
                            <Input
                              id="threshold"
                              type="number"
                              placeholder="e.g., 80"
                              value={alertForm.threshold}
                              onChange={(e) => setAlertForm({ ...alertForm, threshold: e.target.value })}
                              className="w-28 pl-3 pr-6"
                            />
                            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground">%</span>
                          </div>

                          <span className="font-medium">for</span>

                          {/* Time Window Dropdown */}
                          <Select value={alertForm.time_window} onValueChange={(value) => setAlertForm({ ...alertForm, time_window: value })}>
                            <SelectTrigger className="w-[180px]">
                              <SelectValue placeholder="Select duration" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="5 minutes">5 minutes</SelectItem>
                              <SelectItem value="15 minutes">15 minutes</SelectItem>
                              <SelectItem value="30 minutes">30 minutes</SelectItem>
                              <SelectItem value="1 hour">1 hour</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>

                      {/* Section for defining the alert action */}
                      <div className="space-y-4">
                        <h3 className="text-lg font-semibold">Set Notification</h3>
                        <div className="flex items-center gap-3">
                          <span className="font-medium">Severity level:</span>
                          <Select value={alertForm.severity} onValueChange={(value) => setAlertForm({ ...alertForm, severity: value })}>
                            <SelectTrigger className="w-[180px]">
                              <SelectValue placeholder="Select severity" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="info">
                                <div className="flex items-center gap-2">
                                  <div className="w-2 h-2 rounded-full bg-blue-500"></div>
                                  <span>Info</span>
                                </div>
                              </SelectItem>
                              <SelectItem value="warning">
                                <div className="flex items-center gap-2">
                                  <div className="w-2 h-2 rounded-full bg-yellow-500"></div>
                                  <span>Warning</span>
                                </div>
                              </SelectItem>
                              <SelectItem value="critical">
                                <div className="flex items-center gap-2">
                                  <div className="w-2 h-2 rounded-full bg-red-500"></div>
                                  <span>Critical</span>
                                </div>
                              </SelectItem>
                            </SelectContent>
                          </Select>
                        </div>
                      </div>
                    </div>
                  </div>

                  <Separator />

                  {/* Alert Methods */}
                  <div className="space-y-4">
                    <Label className="text-lg font-semibold">Alert Methods</Label>

                    <Button
                      variant={selectedMethods.webhook ? "default" : "outline"}
                      size="sm"
                      className="ml-2"
                      onClick={() => setSelectedMethods({ ...selectedMethods, webhook: !selectedMethods.webhook })}
                    >
                      <Webhook /> webhook
                    </Button>

                    <Button
                      variant={selectedMethods.email ? "default" : "outline"}
                      size="sm"
                      className="ml-2"
                      onClick={() => setSelectedMethods({ ...selectedMethods, email: !selectedMethods.email })}
                    >
                      {" "}
                      <Mail />
                      Email
                    </Button>

                    <Button variant="outline" size="sm" className="ml-2" disabled>
                      <img
                        src={`https://img.icons8.com/?size=100&id=UeC6VYim0wNA&format=png&color=${theme === "dark" ? "#fff" : "#000"}`}
                        alt="Delete"
                        className="h-4 w-4 mr-1"
                      />{" "}
                      Teams Channel
                    </Button>
                    <Button variant="outline" disabled size="sm" className="ml-2">
                      <SlackIcon /> Slack Channel
                    </Button>

                    {selectedMethods.email && (
                      <>
                        {/* Email Selection */}
                        <div className="space-y-3 mt-4 ml-4">
                          <div className="flex items-center">
                            <Dialog open={showEmailDialog} onOpenChange={setShowEmailDialog}>
                              <DialogTrigger asChild>
                                <Button variant="outline" size="sm">
                                  <Plus /> New Email
                                </Button>
                              </DialogTrigger>
                              <DialogContent>
                                <DialogHeader>
                                  <DialogTitle>Add Email for Alerts</DialogTitle>
                                </DialogHeader>
                                <div className="space-y-4">
                                  <div>
                                    <Input
                                      id="email"
                                      type="email"
                                      placeholder="Enter email address"
                                      value={emailForm.email}
                                      className={"mt-3"}
                                      onChange={(e) => setEmailForm({ ...emailForm, email: e.target.value })}
                                    />
                                  </div>
                                  <Button onClick={handleCreateEmail} disabled={creatingEmail}>
                                    {creatingEmail ? "Creating..." : "Create Email"}
                                  </Button>
                                </div>
                              </DialogContent>
                            </Dialog>
                          </div>

                          {verifiedEmails.length > 0 && (
                            <div className="space-y-2 mt-2">
                              {verifiedEmails.map((emailObj, index) => (
                                <div key={index} className="flex items-center space-x-2">
                                  <RadioGroup value={selectedEmails} onValueChange={setSelectedEmails}>
                                    <div key={index} className="flex items-center space-x-2">
                                      <RadioGroupItem value={emailObj.email} id={`email-${index}`} />
                                      <Label htmlFor={`email-${index}`} className="text-sm">
                                        {emailObj.email}
                                      </Label>
                                    </div>
                                  </RadioGroup>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      </>
                    )}
                    {selectedMethods.webhook && (
                      <div className="space-y-3">
                        <Label className="flex items-center space-x-2">
                          <Webhook className="h-4 w-4" />
                          <span>Webhook URL (Optional)</span>
                        </Label>
                        <Input placeholder="https://your-webhook-url.com" value={webhookUrl} onChange={(e) => setWebhookUrl(e.target.value)} />
                      </div>
                    )}
                    {/* Webhook */}
                  </div>

                  <Button
                    onClick={handleCreateAlert}
                    disabled={creatingAlert || !alertForm.threshold || (selectedEmails.length === 0 && !webhookUrl)}
                    className="w-full"
                  >
                    {creatingAlert ? "Creating Alert..." : "Create Alert Rule"}
                  </Button>
                </>
              )}
            </CardContent>
          </Card>

          {/* OTP Verification Dialog */}
          <Dialog open={showOtpDialog} onOpenChange={setShowOtpDialog}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Verify Email</DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <Alert>
                  <Info className="h-4 w-4" />
                  <AlertDescription>A verification code has been sent to {otpForm.email}</AlertDescription>
                </Alert>
                <div>
                  <Label htmlFor="otp">Verification Code</Label>
                  <Input
                    id="otp"
                    placeholder="Enter OTP"
                    className={"mt-3"}
                    value={otpForm.otp}
                    onChange={(e) => setOtpForm({ ...otpForm, otp: e.target.value })}
                  />
                </div>
                <Button onClick={handleVerifyEmail} disabled={verifyingEmail}>
                  {verifyingEmail ? "Verifying..." : "Verify Email"}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </TabsContent>

        <TabsContent value="manage" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Active Alert Rules</CardTitle>
            </CardHeader>
            <CardContent>
              {alertRulesLoading ? (
                <div className="flex justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                </div>
              ) : alertRules.length === 0 ? (
                <div className="text-center py-8 text-gray-500">No alert rules created yet</div>
              ) : (
                <div className="space-y-4">
                  {alertRules.map((rule) => (
                    <Card key={rule.id} className="p-4">
                      <div className="flex items-center justify-between">
                        <div className="space-y-2">
                          <div className="flex items-center space-x-2">
                            <Badge className={`${getSeverityColor(rule.severity)} text-white`}>{rule.severity.toUpperCase()}</Badge>
                            <span className="font-medium">
                              {rule.rule_type === "metric_avg"
                                ? `${rule.metric_name.toUpperCase()} Monitoring`
                                : rule.rule_type === "log_count"
                                ? "Response Status Monitoring"
                                : rule.rule_type === "event_count"
                                ? "Server Restart Detection"
                                : "Unknown"}
                            </span>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            {rule.rule_type === "metric_avg"
                              ? `${rule.metric_name} ${rule.operator} ${rule.threshold}`
                              : rule.rule_type === "log_count"
                              ? `${rule.log_field} ${rule.operator} ${rule.threshold}`
                              : rule.rule_type === "event_count"
                              ? `${rule.event_type} ${rule.operator} ${rule.threshold}`
                              : ""}{" "}
                            over {rule.time_window}
                          </p>
                          <p className="text-xs text-muted-foreground">Created: {new Date(rule.created_at).toLocaleString()}</p>
                        </div>
                        <div className="flex items-center space-x-2">
                          <Badge variant={rule.status === "active" ? "default" : "secondary"}>{rule.status}</Badge>
                        </div>
                      </div>
                    </Card>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default AlertManagement;
