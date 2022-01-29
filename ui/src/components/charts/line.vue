<template>
    <div>
        <div :id=id style="width: 100%; height: 300px;"></div>
    </div>
</template>

<script>
    import Highcharts from 'highcharts';
    export default {
        mounted() {
            this.init();
        },
        data() {
            return {}
        },
        props: {
            id: {
                type: String,
                default: "charts"
            }
        },
        methods: {
            init() {
                this.draw();
            },
            draw() {
                new Highcharts.chart(this.id, {
                    title: {
                        text: ''
                    },
                    yAxis: {
                        title: {
                            text: '单位:Gi/核'
                        }
                    },
                    xAxis: {
                        type: 'datetime',
                        title: {
                            text: null
                        }
                    },
                    legend: {
                        layout: 'vertical',
                        align: 'right',
                        verticalAlign: 'bottom'
                    },
                    credits: { enabled: false },
                    plotOptions: {
                        series: {
                            label: {
                                connectorAllowed: false
                            },
                            pointStart: 0
                        }
                    },
                    series: [{
                        name: 'Memory',
                        data: [
                            [Date.UTC(2020,7,24,12,1,1), 2.2], 
                            [Date.UTC(2020,7,24,12,2,1), 2.5],
                            [Date.UTC(2020,7,24,12,5,1), 2.1], 
                            [Date.UTC(2020,7,24,12,8,1), 2.9], 
                            [Date.UTC(2020,7,24,12,9,1), 2.6]
                        ]
                    }, {
                        name: 'CPU',
                        data: [
                            [Date.UTC(2020,7,24,12,3,1), 2],
                            [Date.UTC(2020,7,24,12,5,1), 5],
                            [Date.UTC(2020,7,24,12,6,1), 4],
                            [Date.UTC(2020,7,24,12,8,1), 2],
                            [Date.UTC(2020,7,24,12,11,1), 3]
                        ]
                    }],
                    responsive: {
                        rules: [{
                            condition: {
                                maxWidth: 500
                            },
                            chartOptions: {
                                legend: {
                                    layout: 'horizontal',
                                    align: 'right',
                                    verticalAlign: 'bottom'
                                }
                            }
                        }]
                    }
                });
            }
        }
    }
</script>